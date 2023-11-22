package data

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"

	"github.com/gocql/gocql"
)

type ReservationRepo struct {
	session *gocql.Session
	logger  *log.Logger
}

// Constructor
func New(logger *log.Logger, session *gocql.Session) (*ReservationRepo, error) {
	return &ReservationRepo{
		session: session,
		logger:  logger,
	}, nil
}

// Disconnect
func (rr *ReservationRepo) CloseSession() {
	rr.session.Close()
}

func (rr *ReservationRepo) CreateTables() error {
	err := rr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
        (id UUID, id_accommodation TEXT, start_date TIMESTAMP, end_date TIMESTAMP, 
        price DOUBLE, price_per_guest BOOLEAN, 
        PRIMARY KEY ((id_accommodation), id)) 
        WITH CLUSTERING ORDER BY (id DESC)`,
			"available_periods_by_accommodation")).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	err = rr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
        (id UUID, id_accommodation TEXT, id_available_period UUID, id_user TEXT,
        start_date TIMESTAMP, end_date TIMESTAMP, guest_number INT, price DOUBLE,
        PRIMARY KEY ((id_available_period), start_date, id)) 
        WITH CLUSTERING ORDER BY (start_date ASC, id DESC)`,
			"reservations_by_available_period")).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

func (rr *ReservationRepo) GetAvailablePeriodsByAccommodation(id string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`
		SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
		FROM available_periods_by_accommodation WHERE id_accommodation = ?`,
		id).Iter().Scanner()

	var availablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var period AvailablePeriodByAccommodation
		var idAccommodation string

		err := scanner.Scan(&period.ID, &idAccommodation, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}

		// Convert idAccommodation string to primitive.ObjectID
		objectID, err := primitive.ObjectIDFromHex(idAccommodation)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}
		period.IDAccommodation = objectID

		availablePeriods = append(availablePeriods, &period)
	}
	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return availablePeriods, nil
}

func (rr *ReservationRepo) GetReservationsByAvailablePeriod(idAvailablePeriod string) (Reservations, error) {
	scanner := rr.session.Query(`
		SELECT id, id_accommodation, id_available_period, id_user, start_date, end_date, guest_number, price
		FROM reservations_by_available_period WHERE id_available_period = ?`,
		idAvailablePeriod).Iter().Scanner()

	var reservations Reservations
	for scanner.Next() {
		var reservation ReservationByAvailablePeriod
		var idAccommodationStr, idUserStr string

		err := scanner.Scan(&reservation.ID, &idAccommodationStr, &reservation.IDAvailablePeriod, &idUserStr, &reservation.StartDate, &reservation.EndDate, &reservation.GuestNumber, &reservation.Price)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}

		// Convert idAccommodationStr and idUserStr strings to primitive.ObjectID
		idAccommodation, err := primitive.ObjectIDFromHex(idAccommodationStr)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}
		reservation.IDAccommodation = idAccommodation

		idUser, err := primitive.ObjectIDFromHex(idUserStr)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}
		reservation.IDUser = idUser

		reservations = append(reservations, &reservation)
	}
	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return reservations, nil
}

func (rr *ReservationRepo) InsertAvailablePeriodByAccommodation(availablePeriod *AvailablePeriodByAccommodation) error {
	var err error
	if availablePeriod.Price < 0 {
		err = errors.New("price cannot be negative")
		return err
	}

	if availablePeriod.StartDate.Before(time.Now()) {
		err = errors.New("start date must be in the future")
		return err
	}

	if availablePeriod.StartDate.After(availablePeriod.EndDate) {
		err = errors.New("start date must be before end date")
		return err
	}

	isOverLap, err := rr.checkForOverlap(*availablePeriod, availablePeriod.IDAccommodation.String())
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	if isOverLap {
		err = errors.New("date overlap")
		return err
	}

	availablePeriodId, _ := gocql.RandomUUID()
	idAccommodation := availablePeriod.IDAccommodation.Hex()
	err = rr.session.Query(
		`INSERT INTO available_periods_by_accommodation (id, id_accommodation, start_date, end_date, price, price_per_guest) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		availablePeriodId, idAccommodation, availablePeriod.StartDate, availablePeriod.EndDate,
		availablePeriod.Price, availablePeriod.PricePerGuest).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

func (rr *ReservationRepo) InsertReservationByAvailablePeriod(reservation *ReservationByAvailablePeriod) error {
	reservationId, _ := gocql.RandomUUID()

	// Convert primitive.ObjectID to string
	idAccommodationStr := reservation.IDAccommodation.Hex()
	idUserStr := reservation.IDUser.Hex()

	err := rr.session.Query(
		`INSERT INTO reservations_by_available_period (id, id_accommodation, id_available_period, id_user, start_date, end_date, guest_number, price) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		reservationId, idAccommodationStr, reservation.IDAvailablePeriod, idUserStr,
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
	availablePeriods, err := rr.FindAvailablePeriodById(id.String(), accommodationdId.String())
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	if len(availablePeriods) != 1 {
		err = errors.New("invalid id")
		return err
	}

	reservations, err := rr.GetReservationsByAvailablePeriod(id.String())
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	if len(reservations) != 0 {
		err = errors.New("cannot change period with reservations")
		return err
	}

	if availablePeriod.Price < 0 {
		err = errors.New("price cannot be negative")
		return err
	}

	if availablePeriod.StartDate.Before(time.Now()) {
		err = errors.New("start date must be in the future")
		return err
	}

	if availablePeriod.StartDate.After(availablePeriod.EndDate) {
		err = errors.New("start date must be before end date")
		return err
	}

	err = rr.session.Query(
		`UPDATE available_periods_by_accommodation 
		SET  end_date = ?, price = ?, price_per_guest = ?, start_date = ? 
		WHERE id = ? AND id_accommodation = ?`,
		availablePeriod.EndDate, availablePeriod.Price, availablePeriod.PricePerGuest,
		availablePeriod.StartDate, availablePeriod.ID, availablePeriod.IDAccommodation).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	return nil
}

func (rr *ReservationRepo) FindAvailablePeriodsByAccommodationId(accommodationId string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
			FROM available_periods_by_accommodation WHERE id_accommodation = ?`, accommodationId).Iter().Scanner()

	var availablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var period AvailablePeriodByAccommodation
		err := scanner.Scan(&period.ID, &period.IDAccommodation, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)
		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}
		availablePeriods = append(availablePeriods, &period)
	}
	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}
	return availablePeriods, nil
}

func (rr *ReservationRepo) FindAvailablePeriodById(id, accommodationId string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
			FROM available_periods_by_accommodation WHERE id = ? AND id_accommodation = ?`,
		id, accommodationId).Iter().Scanner()

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

func (rr *ReservationRepo) FindAllReservationsByAvailablePeriod(periodId string) (Reservations, error) {
	scanner := rr.session.Query(`SELECT id, id_accommodation, id_available_period, id_user, start_date, 
       								    end_date, guest_number, price FROM reservations_by_available_period 
       		                              WHERE id_available_period = ?`, periodId).Iter().Scanner()

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

func (rr *ReservationRepo) checkForOverlap(newPeriod AvailablePeriodByAccommodation, accommodationId string) (bool, error) {
	avalablePeriods, err := rr.FindAvailablePeriodsByAccommodationId(accommodationId)
	if err != nil {
		rr.logger.Println(err)
		return true, err
	}

	// Check for overlaps within the reserved periods of the found reservation.
	if len(avalablePeriods) == 0 {
		return false, nil
	}

	for _, existingPeriod := range avalablePeriods {
		var result bool
		if existingPeriod.ID != newPeriod.ID {
			result = rr.isAvailablePeriodOverlap(*existingPeriod, newPeriod)
			if result {
				return true, nil // Overlap found
			}
		}
	}

	return false, nil // No overlap found
}

func (rr *ReservationRepo) isAvailablePeriodOverlap(currentPeriod AvailablePeriodByAccommodation, newPeriod AvailablePeriodByAccommodation) bool {
	if (newPeriod.StartDate.After(currentPeriod.StartDate) && newPeriod.StartDate.Before(currentPeriod.EndDate)) ||
		(newPeriod.EndDate.After(currentPeriod.StartDate) && newPeriod.EndDate.Before(currentPeriod.EndDate)) ||
		(currentPeriod.StartDate.After(newPeriod.StartDate) && currentPeriod.StartDate.Before(newPeriod.EndDate)) ||
		(currentPeriod.EndDate.After(newPeriod.StartDate) && currentPeriod.EndDate.Before(newPeriod.EndDate)) ||
		currentPeriod.EndDate.Equal(newPeriod.EndDate) {
		return true
	}
	return false
}

func (rr *ReservationRepo) calculatePrice(price float64, pricePerGuest bool, startDate, endDate time.Time, numberOfGuest int16) float64 {
	dateDifference := endDate.Sub(startDate)

	daysDifference := float64(dateDifference.Hours() / 24)

	if pricePerGuest {
		return daysDifference * price * float64(numberOfGuest)
	}

	return daysDifference * price
}

func (rr *ReservationRepo) convertObjectIDToUUID(objectID primitive.ObjectID) (gocql.UUID, error) {
	// Konvertujte ObjectID u heksadecimalni string
	hexString := objectID.Hex()

	// Parsirajte heksadecimalni string u gocql.UUID
	uuid, err := gocql.ParseUUID(hexString)
	if err != nil {
		return gocql.UUID{}, err
	}

	return uuid, nil
}
