package data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	// NoSQL: module containing Mongo api client
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// TODO "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type ReservationRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// Constructor
func New(ctx context.Context, logger *log.Logger) (*ReservationRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &ReservationRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect
func (rr *ReservationRepo) Disconnect(ctx context.Context) error {
	err := rr.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Check database connection
func (rr *ReservationRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := rr.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		rr.logger.Println(err)
	}

	// Print available databases
	databases, err := rr.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
	}
	fmt.Println(databases)
}

func (rr *ReservationRepo) GetAll() (Reservations, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reservationCollection := rr.getCollection()

	var reservations Reservations
	reserationCursor, err := reservationCollection.Find(ctx, bson.M{})
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	if err = reserationCursor.All(ctx, &reservations); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return reservations, nil
}

func (rr *ReservationRepo) GetReservationById(id string) (*Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reservationCollection := rr.getCollection()

	var reservation Reservation
	objID, _ := primitive.ObjectIDFromHex(id)
	err := reservationCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&reservation)
	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return &reservation, nil
}

func (rr *ReservationRepo) GetReservedPeriodById(reservationID, periodID string) (*ReservedPeriod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservations := rr.getCollection()

	objReservationID, err := primitive.ObjectIDFromHex(reservationID)
	if err != nil {
		return nil, err
	}

	periodObjId, _ := primitive.ObjectIDFromHex(periodID)

	filter := bson.M{
		"_id":                     objReservationID,
		"availabilityPeriods._id": periodObjId,
	}

	var reservation Reservation
	err = reservations.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("Reservation not found")
		}
		return nil, err
	}

	// Find the specific AvailabilityPeriod within the Reservation.
	for _, period := range *reservation.ReservedPeriods {
		if period.ID.Hex() == periodID {
			return &period, nil
		}
	}

	return nil, errors.New("AvailabilityPeriod not found in Reservation")
}

func (rr *ReservationRepo) PostReservation(reservation *Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservationCollection := rr.getCollection()

	reservation.ID = primitive.NewObjectID()
	reservation.ReservedPeriods = &[]ReservedPeriod{}

	if reservation.PricePerAccommodationConfiguration.StandardPrice < 0 ||
		reservation.PricePerAccommodationConfiguration.SummerSeasonPrice < 0 ||
		reservation.PricePerAccommodationConfiguration.WinterSeasonPrice < 0 ||
		reservation.PricePerAccommodationConfiguration.WeekendSeasonPrice < 0 {
		rr.logger.Println("Price per reservation cannot be negative")
		return nil
	}

	if reservation.PricePerGuestConfiguration.StandardPrice < 0 ||
		reservation.PricePerGuestConfiguration.SummerSeasonPrice < 0 ||
		reservation.PricePerGuestConfiguration.WinterSeasonPrice < 0 ||
		reservation.PricePerGuestConfiguration.WeekendSeasonPrice < 0 {
		rr.logger.Println("Price per guest cannot be negative")
		return nil
	}

	result, err := reservationCollection.InsertOne(ctx, &reservation)

	if err != nil {
		rr.logger.Println(err)
		return err
	}
	rr.logger.Printf("Reservation Id: %v\n", result.InsertedID)
	return nil
}

// Delete this method
//func (rr *ReservationRepo) CreateReservation(ctx context.Context, reservationID primitive.ObjectID, periodID primitive.ObjectID) error {
//	reservationCollection := rr.getCollection()
//	filter := bson.M{"_id": reservationID, "availabilityPeriods._id": periodID}
//	update := bson.M{
//		"$set": bson.M{"availabilityPeriods.$.isAvailable": true},
//	}
//	updateOptions := options.Update().SetUpsert(false)
//
//	_, err := reservationCollection.UpdateOne(ctx, filter, update, updateOptions)
//	if err != nil {
//		log.Printf("Failed to reserve availability period: %v\n", err)
//		return err
//	}
//
//	return nil
//}

func (rr *ReservationRepo) AddReservedPeriod(id string, period *ReservedPeriod) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservations := rr.getCollection()

	objId, _ := primitive.ObjectIDFromHex(id)
	period.ID = primitive.NewObjectID()

	isNotValidPeriod, _ := rr.checkForOverlap(objId, *period)
	if isNotValidPeriod {
		rr.logger.Println("Not valid period date")
		return nil
	}

	period.Price = rr.calculatePrice(id, period)

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}

	update := bson.M{"$push": bson.M{
		"reservedPeriods": period,
	}}
	result, err := reservations.UpdateOne(ctx, filter, update)
	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

// TODO Stevan 1.9
func (rr *ReservationRepo) UpdateReservedPeriod(reservationId string, newPeriod *ReservedPeriod) {
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	//reservations := rr.getCollection()
	//
	//objId, _ := primitive.ObjectIDFromHex(reservationId)
	//periodObjId := newPeriod.ID.Hex()
	//
	//period, _ := rr.FindAvailablePeriodById(reservationId, periodObjId)
	//
	//if period != nil {
	//	if newPeriod.PriceConfiguration.PricePerAccommodation > 0 {
	//		period.PriceConfiguration.PricePerAccommodation = newPeriod.PriceConfiguration.PricePerAccommodation
	//	}
	//	if newPeriod.PriceConfiguration.PricePerGuest > 0 {
	//		period.PriceConfiguration.PricePerGuest = newPeriod.PriceConfiguration.PricePerGuest
	//	}
	//	if !newPeriod.StartDate.IsZero() && newPeriod.StartDate.After(time.Now()) && newPeriod.StartDate.Before(period.EndDate) {
	//		period.StartDate = newPeriod.StartDate
	//	}
	//	if !newPeriod.EndDate.IsZero() && newPeriod.EndDate.After(time.Now()) && newPeriod.EndDate.After(period.StartDate) {
	//		period.EndDate = newPeriod.EndDate
	//	}
	//	period.PriceConfiguration.UsePricePerGuest = newPeriod.PriceConfiguration.UsePricePerGuest
	//
	//	filter := bson.M{"_id": objId}
	//	update := bson.M{
	//		"$set": bson.M{
	//			"availabilityPeriods": []*AvailabilityPeriod{period},
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
	//}

}

func (rr *ReservationRepo) DeleteReservationById(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservationCollection := rr.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := reservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	rr.logger.Printf("Reservation deleted: %v\n", result.DeletedCount)
	return nil
}

//Delete this method
//func (rr *ReservationRepo) ReservePeriod(reservationId string, periodId string) error {
//	reservations := rr.getCollection()
//
//	reservationObjID, _ := primitive.ObjectIDFromHex(reservationId)
//	availabilityPeriodObjID, _ := primitive.ObjectIDFromHex(periodId)
//
//	filter := bson.M{
//		"_id":                     reservationObjID,
//		"availabilityPeriods._id": availabilityPeriodObjID,
//	}
//	update := bson.M{
//		"$set": bson.M{
//			"availabilityPeriods.$.isAvailable": false,
//		},
//	}
//
//	result, err := reservations.UpdateOne(context.Background(), filter, update)
//
//	rr.logger.Printf("Documents matched: %v\n", result.MatchedCount)
//	rr.logger.Printf("Documents updated: %v\n", result.ModifiedCount)
//
//	if err != nil {
//		log.Fatal(err)
//		return err
//	}
//
//	return nil
//}

func (rr *ReservationRepo) getCollection() *mongo.Collection {
	patientDatabase := rr.cli.Database("reservationDB")
	patientCollection := patientDatabase.Collection("reservation")
	return patientCollection
}

func (rr *ReservationRepo) checkForOverlap(reservationID primitive.ObjectID, newPeriod ReservedPeriod) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservations := rr.getCollection()

	// Find the reservation by its ID.
	filter := bson.M{"_id": reservationID}
	var foundReservation Reservation
	err := reservations.FindOne(ctx, filter).Decode(&foundReservation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			rr.logger.Println(err)
			return false, nil
		}
		return false, err
	}

	// Check for overlaps within the reserved periods of the found reservation.
	if len(*foundReservation.ReservedPeriods) == 0 {
		return false, nil
	}
	for _, existingPeriod := range *foundReservation.ReservedPeriods {
		var result bool
		result = rr.isPeriodOverlap(existingPeriod, newPeriod)
		if result {
			return true, nil // Overlap found
		}
	}

	return false, nil // No overlap found
}

func (rr *ReservationRepo) isPeriodOverlap(curentPeriod ReservedPeriod, newPeriod ReservedPeriod) bool {
	if newPeriod.StartDate.After(time.Now()) &&
		newPeriod.StartDate.Before(newPeriod.EndDate) &&
		newPeriod.StartDate.After(curentPeriod.StartDate) &&
		newPeriod.StartDate.After(curentPeriod.EndDate) {
		return false
	} else {
		return true
	}
}

func (rr *ReservationRepo) calculatePrice(reservationId string, period *ReservedPeriod) float64 {
	reservation, _ := rr.GetReservationById(reservationId)
	isPricePerGuest := reservation.PricePerGuest

	year := period.StartDate.Year()

	startOfSummerSeason := time.Date(year, time.June, 1, 0, 0, 0, 0, time.UTC)
	endOfSummerSeason := time.Date(year, time.September, 1, 0, 0, 0, 0, time.UTC)

	if period.StartDate.After(startOfSummerSeason) && period.StartDate.Before(endOfSummerSeason) {
		if !isPricePerGuest {
			return reservation.PricePerAccommodationConfiguration.SummerSeasonPrice
		}
		return reservation.PricePerGuestConfiguration.SummerSeasonPrice * float64(period.NumberOfGuests)
	}

	startOfWinterSeason := time.Date(year, time.December, 1, 0, 0, 0, 0, time.UTC)
	endOfWinterSeason := time.Date(year, time.February, 1, 0, 0, 0, 0, time.UTC)

	if period.StartDate.After(startOfWinterSeason) && period.StartDate.Before(endOfWinterSeason) {
		if !isPricePerGuest {
			return reservation.PricePerAccommodationConfiguration.WinterSeasonPrice
		}
		return reservation.PricePerGuestConfiguration.WinterSeasonPrice * float64(period.NumberOfGuests)
	}

	isWeekend := period.StartDate.Weekday().String()
	if isWeekend == "sunday" || isWeekend == "saturday" {
		if !isPricePerGuest {
			return reservation.PricePerAccommodationConfiguration.WeekendSeasonPrice
		}
		return reservation.PricePerGuestConfiguration.WeekendSeasonPrice * float64(period.NumberOfGuests)
	}

	if isPricePerGuest {
		return reservation.PricePerGuestConfiguration.StandardPrice * float64(period.NumberOfGuests)
	}

	return reservation.PricePerAccommodationConfiguration.StandardPrice * float64(period.NumberOfGuests)
}
