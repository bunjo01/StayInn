package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"os"
	"time"
)

type ReservationRepo struct {
	session *gocql.Session
	logger  *log.Logger
}

// Constructor
func New(ctx context.Context, logger *log.Logger) (*ReservationRepo, error) {
	db := os.Getenv("CASS_DB")

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy'.
						'replication_factor': %d
					}`, "reservation", 1)).Exec()
	if err != nil {
		logger.Println(err)
	}

	session.Close()

	cluster.Keyspace = "reservation"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	return &ReservationRepo{
		session: session,
		logger:  logger,
	}, nil
}

// Disconnect
func (rr *ReservationRepo) CloseSession() {
	rr.session.Close()
}

func (rr *ReservationRepo) CreateTables() {
	err := rr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
        (id UUID, id_accommodation UUID, start_date TIMESTAMP, end_date TIMESTAMP, 
        price DOUBLE, price_per_guest BOOLEAN, 
        PRIMARY KEY ((id_accommodation), start_date, id)) 
        WITH CLUSTERING ORDER BY (start_date ASC, id DESC)`,
			"available_periods_by_accommodation")).Exec()
	if err != nil {
		rr.logger.Println(err)
	}

	err = rr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
        (id UUID, id_accommodation UUID, id_available_period UUID, id_user UUID,
        start_date TIMESTAMP, end_date TIMESTAMP, guest_number INT, price DOUBLE,
        PRIMARY KEY ((id_available_period), start_date, id)) 
        WITH CLUSTERING ORDER BY (start_date ASC, id DESC)`,
			"reservations_by_available_period")).Exec()
	if err != nil {
		rr.logger.Println(err)
	}

}

func (rr *ReservationRepo) GetAvailablePeriodsByAccommodation(id string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
			FROM available_periods_by_accommodation WHERE id_accommodation = ?`,
		id).Iter().Scanner()

	var avaiablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var period AvailablePeriodByAccommodation
		err := scanner.Scan(&period.ID, &period.IDAccommodation, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}
		avaiablePeriods = append(avaiablePeriods, &period)
	}
	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return avaiablePeriods, nil
}

func (rr *ReservationRepo) GetReservationsByAvailablePeriod(id string) (Reservations, error) {
	scanner := rr.session.Query(`SELECT id, id_accommodation, id_available_period, id_user, start_date, 
       		end_date, guest_number, price 
			FROM reservations_by_available_period WHERE id_available_period = ?`,
		id).Iter().Scanner()

	var reservations Reservations
	for scanner.Next() {
		var reservation ReservationByAvailablePeriod
		err := scanner.Scan(&reservation.ID, &reservation.IDAccommodation, &reservation.IDAvailablePeriod, &reservation.IDUser,
			&reservation.StartDate, &reservation.EndDate, &reservation.GuestNumber, &reservation.Price)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}
		reservations = append(reservations, &reservation)
	}
	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return reservations, nil
}

func (rr *ReservationRepo) InsertAvailablePeriodByAccommodation(availablePeriod *AvailablePeriodByAccommodation) error {
	availablePeriodId, _ := gocql.RandomUUID()
	err := rr.session.Query(
		`INSERT INTO available_periods_by_accommodation (id, id_accommodation, start_date, end_date, price, price_per_guest) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		availablePeriodId, availablePeriod.IDAccommodation, availablePeriod.StartDate, availablePeriod.EndDate,
		availablePeriod.Price, availablePeriod.PricePerGuest).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

func (rr *ReservationRepo) InsertReservationByAvailablePeriod(reservation *ReservationByAvailablePeriod) error {
	reservationId, _ := gocql.RandomUUID()
	err := rr.session.Query(
		`INSERT INTO reservations_by_available_period (id, id_accommodation, id_available_period, id_user, start_date, end_date, guest_number, price) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		reservationId, reservation.IDAccommodation, reservation.IDAvailablePeriod, reservation.IDUser,
		reservation.StartDate, reservation.EndDate, reservation.GuestNumber, reservation.Price).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

func (rr *ReservationRepo) UpdateAvailablePeriodByAccommodation(availablePeriod *AvailablePeriodByAccommodation) error {
	id := availablePeriod.ID
	accommodationdId := availablePeriod.IDAccommodation
	availablePeriods, err := rr.FindAvailablePeriodById(id.String(), accommodationdId.String(), availablePeriod.StartDate)
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	if len(availablePeriods) != 1 {
		err = errors.New("invalid id")
		return err
	}

	err = rr.session.Query(
		`UPDATE available_periods_by_accommodation 
		SET  end_date = ?, price = ?, price_per_guest = ? 
		WHERE id = ? AND id_accommodation = ? AND start_date = ?`,
		availablePeriod.EndDate,
		availablePeriod.Price, availablePeriod.PricePerGuest,
		availablePeriod.ID, availablePeriod.IDAccommodation, availablePeriod.StartDate).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	return nil
}

func (rr *ReservationRepo) FindAvailablePeriodById(id, accommodationId string, startDate time.Time) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
			FROM available_periods_by_accommodation WHERE id = ? AND start_date = ? AND id_accommodation = ?`,
		id, startDate, accommodationId).Iter().Scanner()

	var avaiablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var period AvailablePeriodByAccommodation
		err := scanner.Scan(&period.ID, &period.IDAccommodation, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}
		avaiablePeriods = append(avaiablePeriods, &period)
	}
	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return avaiablePeriods, nil
}

func (rr *ReservationRepo) GetDistinctIds(idColumnName string, tableName string) ([]string, error) {
	scanner := rr.session.Query(
		fmt.Sprintf(`SELECT DISTINCT %s FROM %s`, idColumnName, tableName)).Iter().Scanner()
	var ids []string
	for scanner.Next() {
		var id string
		err := scanner.Scan(&id)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return ids, nil
}

//func (rr *ReservationRepo) checkForOverlap(reservationID primitive.ObjectID, newPeriod ReservedPeriod) (bool, error) {
//
//	// Check for overlaps within the reserved periods of the found reservation.
//	if len(*foundReservation.ReservedPeriods) == 0 {
//		return false, nil
//	}
//	for _, existingPeriod := range *foundReservation.ReservedPeriods {
//		var result bool
//		if existingPeriod.ID != newPeriod.ID {
//			result = rr.isPeriodOverlap(existingPeriod, newPeriod)
//			if result {
//				return true, nil // Overlap found
//			}
//		}
//	}
//
//	return false, nil // No overlap found
//}
//
//func (rr *ReservationRepo) isPeriodOverlap(currentPeriod ReservedPeriod, newPeriod ReservedPeriod) bool {
//	if (newPeriod.StartDate.After(currentPeriod.StartDate) && newPeriod.StartDate.Before(currentPeriod.EndDate)) ||
//		(newPeriod.EndDate.After(currentPeriod.StartDate) && newPeriod.EndDate.Before(currentPeriod.EndDate)) ||
//		(currentPeriod.StartDate.After(newPeriod.StartDate) && currentPeriod.StartDate.Before(newPeriod.EndDate)) ||
//		(currentPeriod.EndDate.After(newPeriod.StartDate) && currentPeriod.EndDate.Before(newPeriod.EndDate)) ||
//		currentPeriod.EndDate.Equal(newPeriod.EndDate) {
//		return true
//	}

//
//func (rr *ReservationRepo) GetAll() (Reservations, error) {
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	reservationCollection := rr.getCollection()
//
//	var reservations Reservations
//	reserationCursor, err := reservationCollection.Find(ctx, bson.M{})
//	if err != nil {
//		rr.logger.Println(err)
//		return nil, err
//	}
//	if err = reserationCursor.All(ctx, &reservations); err != nil {
//		rr.logger.Println(err)
//		return nil, err
//	}
//	return reservations, nil
//}
//
//func (rr *ReservationRepo) GetReservationById(id string) (*Reservation, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	reservationCollection := rr.getCollection()
//
//	var reservation Reservation
//	objID, _ := primitive.ObjectIDFromHex(id)
//	err := reservationCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&reservation)
//	if err != nil {
//		rr.logger.Println(err)
//		return nil, err
//	}
//	return &reservation, nil
//}
//
//func (rr *ReservationRepo) GetReservedPeriodById(reservationID, periodID string) (*ReservedPeriod, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	reservations := rr.getCollection()
//
//	objReservationID, err := primitive.ObjectIDFromHex(reservationID)
//	if err != nil {
//		return nil, err
//	}
//
//	periodObjId, _ := primitive.ObjectIDFromHex(periodID)
//
//	filter := bson.M{
//		"_id":                 objReservationID,
//		"reservedPeriods._id": periodObjId,
//	}
//
//	var reservation Reservation
//	err = reservations.FindOne(ctx, filter).Decode(&reservation)
//	if err != nil {
//		if err == mongo.ErrNoDocuments {
//			return nil, errors.New("Reservation not found")
//		}
//		return nil, err
//	}
//
//	// Find the specific AvailabilityPeriod within the Reservation.
//	for _, period := range *reservation.ReservedPeriods {
//		if period.ID.Hex() == periodID {
//			return &period, nil
//		}
//	}
//
//	return nil, errors.New("AvailabilityPeriod not found in Reservation")
//}
//
//func (rr *ReservationRepo) PostReservation(reservation *Reservation) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	reservationCollection := rr.getCollection()
//
//	reservation.ID = primitive.NewObjectID()
//	reservation.ReservedPeriods = &[]ReservedPeriod{}
//
//	isPriceValidError := rr.isPricePerAccommodationValid(reservation)
//	if isPriceValidError != nil {
//		rr.logger.Println(isPriceValidError)
//		return nil
//	}
//
//	isPriceValidError = rr.isPricePerGuestValid(reservation)
//	if isPriceValidError != nil {
//		rr.logger.Println(isPriceValidError)
//		return nil
//	}
//
//	result, err := reservationCollection.InsertOne(ctx, &reservation)
//
//	if err != nil {
//		rr.logger.Println(err)
//		return err
//	}
//	rr.logger.Printf("Reservation Id: %v\n", result.InsertedID)
//	return nil
//}
//
//func (rr *ReservationRepo) AddReservedPeriod(id string, period *ReservedPeriod) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	reservations := rr.getCollection()
//
//	objId, _ := primitive.ObjectIDFromHex(id)
//	period.ID = primitive.NewObjectID()
//
//	isNotValidPeriod, _ := rr.checkForOverlap(objId, *period)
//	if isNotValidPeriod {
//		rr.logger.Println("Not valid period date")
//		return nil
//	}
//
//	period.Price = rr.calculatePrice(id, period)
//
//	filter := bson.D{{Key: "_id", Value: objId}}
//
//	update := bson.M{"$push": bson.M{
//		"reservedPeriods": period,
//	}}
//	result, err := reservations.UpdateOne(ctx, filter, update)
//	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
//	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)
//
//	if err != nil {
//		rr.logger.Println(err)
//		return err
//	}
//	return nil
//}
//
//// TODO : NOT WORKING
//func (rr *ReservationRepo) UpdateReservedPeriod(reservationId string, newPeriod *ReservedPeriod) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	objId, _ := primitive.ObjectIDFromHex(reservationId)
//	periodObjId := newPeriod.ID.Hex()
//
//	period, _ := rr.GetReservedPeriodById(reservationId, periodObjId)
//
//	if period != nil {
//		isOverlap, err := rr.checkForOverlap(objId, *newPeriod)
//		if err != nil {
//			rr.logger.Println(err)
//			return err // Return the error instead of nil
//		}
//
//		if isOverlap {
//			rr.logger.Println("is isOverlap is not null")
//			return nil
//		}
//
//		rr.logger.Println(period.StartDate)
//		rr.logger.Println(period.ID)
//
//		period.StartDate = newPeriod.StartDate
//		period.EndDate = newPeriod.EndDate
//		period.NumberOfGuests = newPeriod.NumberOfGuests
//		period.Price = rr.calculatePrice(reservationId, period)
//
//		filter := bson.M{"_id": objId, "reservedPeriods._id": period.ID}
//		update := bson.M{
//			"$set": bson.M{
//				"reservedPeriods.$.StartDate":      period.StartDate,
//				"reservedPeriods.$.EndDate":        period.EndDate,
//				"reservedPeriods.$.NumberOfGuests": period.NumberOfGuests,
//			},
//		}
//
//		result, err := rr.getCollection().UpdateOne(ctx, filter, update)
//		if err != nil {
//			return err
//		}
//
//		rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
//		rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)
//	}
//
//	return nil
//}
//
//func (rr *ReservationRepo) DeleteReservationById(id string) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	reservationCollection := rr.getCollection()
//
//	objID, _ := primitive.ObjectIDFromHex(id)
//	filter := bson.D{{Key: "_id", Value: objID}}
//	result, err := reservationCollection.DeleteOne(ctx, filter)
//	if err != nil {
//		rr.logger.Println(err)
//		return err
//	}
//	rr.logger.Printf("Reservation deleted: %v\n", result.DeletedCount)
//	return nil
//}
//
//func (rr *ReservationRepo) DeleteReservedPeriod(reservationId, periodId string) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	reservations := rr.getCollection()
//
//	objId, _ := primitive.ObjectIDFromHex(reservationId)
//	_, err := rr.GetReservationById(reservationId)
//	if err != nil {
//		rr.logger.Println(err)
//		return err
//	}
//
//	periodObjId, _ := primitive.ObjectIDFromHex(periodId)
//
//	filter := bson.M{"_id": objId}
//	update := bson.M{
//		"$pull": bson.M{
//			"reservedPeriods": bson.M{"_id": periodObjId},
//		},
//	}
//
//	result, err := reservations.UpdateOne(ctx, filter, update)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
//	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)
//
//	return nil
//}
//
//func (rr *ReservationRepo) getCollection() *mongo.Collection {
//	// patientDatabase := rr.cli.Database("reservationDB")
//	// patientCollection := patientDatabase.Collection("reservation")
//	// return patientCollection
//}
//
//	return false
//}
//
//func (rr *ReservationRepo) calculatePrice(reservationId string, period *ReservedPeriod) float64 {
//	reservation, _ := rr.GetReservationById(reservationId)
//	isPricePerGuest := reservation.PricePerGuest
//
//	year := period.StartDate.Year()
//
//	// Calculates summer season
//	startOfSummerSeason := time.Date(year, time.June, 1, 0, 0, 0, 0, time.UTC)
//	endOfSummerSeason := time.Date(year, time.September, 1, 0, 0, 0, 0, time.UTC)
//
//	if period.StartDate.After(startOfSummerSeason) && period.StartDate.Before(endOfSummerSeason) {
//		if !isPricePerGuest {
//			return reservation.PricePerAccommodationConfiguration.SummerSeasonPrice
//		}
//		return reservation.PricePerGuestConfiguration.SummerSeasonPrice * float64(period.NumberOfGuests)
//	}
//
//	// Calculates winter season
//	startOfWinterSeason := time.Date(year, time.December, 1, 0, 0, 0, 0, time.UTC)
//	endOfWinterSeason := time.Date(year, time.February, 1, 0, 0, 0, 0, time.UTC)
//
//	if period.StartDate.After(startOfWinterSeason) && period.StartDate.Before(endOfWinterSeason) {
//		if !isPricePerGuest {
//			return reservation.PricePerAccommodationConfiguration.WinterSeasonPrice
//		}
//		return reservation.PricePerGuestConfiguration.WinterSeasonPrice * float64(period.NumberOfGuests)
//	}
//
//	// Calculate weekend
//	isWeekend := period.StartDate.Weekday().String()
//	if isWeekend == "sunday" || isWeekend == "saturday" {
//		if !isPricePerGuest {
//			return reservation.PricePerAccommodationConfiguration.WeekendSeasonPrice
//		}
//		return reservation.PricePerGuestConfiguration.WeekendSeasonPrice * float64(period.NumberOfGuests)
//	}
//
//	// Calculate standard
//	if isPricePerGuest {
//		return reservation.PricePerGuestConfiguration.StandardPrice * float64(period.NumberOfGuests)
//	}
//
//	return reservation.PricePerAccommodationConfiguration.StandardPrice * float64(period.NumberOfGuests)
//}
//
//func (rr *ReservationRepo) isPricePerAccommodationValid(reservation *Reservation) error {
//	if reservation.PricePerAccommodationConfiguration.StandardPrice < 0 ||
//		reservation.PricePerAccommodationConfiguration.SummerSeasonPrice < 0 ||
//		reservation.PricePerAccommodationConfiguration.WinterSeasonPrice < 0 ||
//		reservation.PricePerAccommodationConfiguration.WeekendSeasonPrice < 0 {
//		return errors.New("price per accommodation cannot be negative")
//	}
//	return nil
//}
//
//func (rr *ReservationRepo) isPricePerGuestValid(reservation *Reservation) error {
//	if reservation.PricePerGuestConfiguration.StandardPrice < 0 ||
//		reservation.PricePerGuestConfiguration.SummerSeasonPrice < 0 ||
//		reservation.PricePerGuestConfiguration.WinterSeasonPrice < 0 ||
//		reservation.PricePerGuestConfiguration.WeekendSeasonPrice < 0 {
//		return errors.New("price per guest cannot be negative")
//	}
//	return nil
//}
