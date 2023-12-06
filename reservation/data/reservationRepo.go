package data

import (
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

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
        PRIMARY KEY ((id_available_period),  id)) 
        WITH CLUSTERING ORDER BY (id ASC)`,
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

	isOverLap, err := rr.checkForOverlap(*availablePeriod, availablePeriod.IDAccommodation.Hex())
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

	// Check if the reservation is within the appropriate range of the available period
	availablePeriod, err := rr.FindAvailablePeriodById(reservation.IDAvailablePeriod.String(), reservation.IDAccommodation.Hex())
	if err != nil {
		rr.logger.Println("Error obtaining available period:", err)
		rr.logger.Println(availablePeriod.ToJSON(log.Writer()))
		return err
	}
	if reservation.StartDate.Before(availablePeriod.StartDate) || reservation.EndDate.After(availablePeriod.EndDate) {
		rr.logger.Println("Reservation is not within the appropriate range of the available period")
		return errors.New("reservation is not within the appropriate range of the available period")
	}

	// Retrieve existing reservations for the available period
	existingReservations, err := rr.FindAllReservationsByAvailablePeriod(availablePeriod.ID.String())
	if err != nil {
		rr.logger.Println("Error obtaining existing reservations:", err)
		return err
	}

	// Check for overlapping reservations
	for _, existingReservation := range existingReservations {
		if (reservation.StartDate.Before(existingReservation.EndDate) || reservation.StartDate.Equal(existingReservation.EndDate)) &&
			(reservation.EndDate.After(existingReservation.StartDate) || reservation.EndDate.Equal(existingReservation.StartDate)) {
			rr.logger.Println("New reservation overlaps with an existing reservation.")
			return errors.New("new reservation overlaps with an existing reservation")
		}
	}

	calculatedPrice := rr.calculatePrice(availablePeriod.Price, availablePeriod.PricePerGuest, reservation.StartDate, reservation.EndDate, int16(reservation.GuestNumber))
	err = rr.session.Query(
		`INSERT INTO reservations_by_available_period 
		(id, id_accommodation, id_available_period, id_user, start_date, end_date, guest_number, price) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		reservationId, reservation.IDAccommodation.Hex(), reservation.IDAvailablePeriod, reservation.IDUser.Hex(),
		reservation.StartDate, reservation.EndDate, reservation.GuestNumber, calculatedPrice).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

func (rr *ReservationRepo) UpdateAvailablePeriodByAccommodation(availablePeriod *AvailablePeriodByAccommodation) error {
	id := availablePeriod.ID
	accommodationdId := availablePeriod.IDAccommodation.Hex()
	availablePeriods, err := rr.FindAvailablePeriodsById(id.String(), accommodationdId)
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

	isOverLap, err := rr.checkForOverlap(*availablePeriod, accommodationdId)
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	if isOverLap {
		err = errors.New("date overlap")
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
		availablePeriod.StartDate, availablePeriod.ID.String(), availablePeriod.IDAccommodation.Hex()).Exec()
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	return nil
}

func (rr *ReservationRepo) FindAvailablePeriodsByAccommodationId(accommodationId string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`
    SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
    FROM available_periods_by_accommodation 
    WHERE id_accommodation = ?`, accommodationId).Iter().Scanner()

	var availablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var (
			idAccommodationStr string
			period             AvailablePeriodByAccommodation
		)

		err := scanner.Scan(&period.ID, &idAccommodationStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)

		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}

		// Convert string to primitive.ObjectID
		period.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)

		availablePeriods = append(availablePeriods, &period)
	}

	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	return availablePeriods, nil
}

func (rr *ReservationRepo) FindAvailablePeriodsById(id, accommodationId string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`
        SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
        FROM available_periods_by_accommodation 
        WHERE id = ? AND id_accommodation = ?`,
		id, accommodationId).Iter().Scanner()

	var availablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var (
			idAccommodationStr string
			period             AvailablePeriodByAccommodation
		)

		err := scanner.Scan(&period.ID, &idAccommodationStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)

		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}

		period.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)

		availablePeriods = append(availablePeriods, &period)
	}

	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	return availablePeriods, nil
}

func (rr *ReservationRepo) FindAvailablePeriodById(id, accommodationID string) (*AvailablePeriodByAccommodation, error) {
	query := `SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
          FROM available_periods_by_accommodation 
          WHERE id = ? AND id_accommodation = ? LIMIT 1`

	var (
		idAccommodationStr string
		period             AvailablePeriodByAccommodation
	)

	err := rr.session.Query(query, id, accommodationID).Consistency(gocql.One).Scan(
		&period.ID, &idAccommodationStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest,
	)

	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	// Convert string to primitive.ObjectID
	period.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)

	return &period, nil
}

func (rr *ReservationRepo) FindAllReservationsByAvailablePeriod(periodId string) (Reservations, error) {
	scanner := rr.session.Query(`
        SELECT id, id_accommodation, id_available_period, id_user, start_date, end_date, guest_number, price
        FROM reservations_by_available_period
        WHERE id_available_period = ?`, periodId).Iter().Scanner()

	var reservations Reservations
	for scanner.Next() {
		var (
			idAccommodationStr string
			idUserStr          string
			reservation        ReservationByAvailablePeriod
		)

		err := scanner.Scan(&reservation.ID, &idAccommodationStr, &reservation.IDAvailablePeriod, &idUserStr,
			&reservation.StartDate, &reservation.EndDate, &reservation.GuestNumber, &reservation.Price)

		if err != nil {
			rr.logger.Println(err)
			return nil, err
		}

		// Convert strings to primitive.ObjectID
		reservation.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
		reservation.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

		reservations = append(reservations, &reservation)
	}

	if err := scanner.Err(); err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	return reservations, nil
}

func (rr *ReservationRepo) FindReservationByIdAndAvailablePeriod(id, periodID string) (*ReservationByAvailablePeriod, error) {
	query := `SELECT id, id_accommodation, id_available_period, id_user, start_date, 
               end_date, guest_number, price 
               FROM reservations_by_available_period 
               WHERE id = ? AND id_available_period = ? LIMIT 1`

	var (
		idAccommodationStr string
		idUserStr          string
		reservation        ReservationByAvailablePeriod
	)

	err := rr.session.Query(query, id, periodID).Consistency(gocql.One).Scan(
		&reservation.ID, &idAccommodationStr, &reservation.IDAvailablePeriod, &idUserStr,
		&reservation.StartDate, &reservation.EndDate, &reservation.GuestNumber, &reservation.Price,
	)

	if err != nil {
		rr.logger.Println(err)
		return nil, err
	}

	// Convert strings to primitive.ObjectID
	reservation.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
	reservation.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

	return &reservation, nil
}

func (rr *ReservationRepo) DeleteReservationByIdAndAvailablePeriodID(id, periodID string) error {
	reservation, err := rr.FindReservationByIdAndAvailablePeriod(id, periodID)
	if err != nil {
		rr.logger.Println(err)
		return err
	}

	// Check if the start date of the reservation has passed
	if time.Now().After(reservation.StartDate) {
		// If the start date has passed, disallow deletion and return an error
		return errors.New("cannot delete reservation after start date has passed")
	}

	query := `DELETE FROM reservations_by_available_period
              WHERE id = ? AND id_available_period = ?`

	if err := rr.session.Query(query, id, periodID).Exec(); err != nil {
		rr.logger.Println(err)
		return err
	}
	return nil
}

// **SEARCH**

func (rr *ReservationRepo) FindAccommodationIdsByDates(dates *Dates) (ListOfObjectIds, error) {
	var periodIDs []gocql.UUID
	uniqueAccommodationIds := make(map[primitive.ObjectID]struct{})

	for _, id := range dates.AccommodationIds {
		scanner := rr.session.Query(`
			SELECT id, id_accommodation, start_date, end_date, price, price_per_guest 
			FROM available_periods_by_accommodation 
			WHERE id_accommodation = ?`, id.Hex()).Iter().Scanner()

		for scanner.Next() {
			var (
				idAccommodationStr string
				period             AvailablePeriodByAccommodation
			)

			err := scanner.Scan(&period.ID, &idAccommodationStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)

			if err != nil {
				rr.logger.Println(err)
				return ListOfObjectIds{}, err
			}

			period.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)

			if (period.StartDate.Before(dates.StartDate) || period.StartDate.Equal(dates.StartDate)) &&
				(period.EndDate.After(dates.EndDate) || period.EndDate.Equal(dates.EndDate)) {
				periodIDs = append(periodIDs, period.ID)
				uniqueAccommodationIds[period.IDAccommodation] = struct{}{}
			}
		}

		if err := scanner.Err(); err != nil {
			rr.logger.Println(err)
			return ListOfObjectIds{}, err
		}
	}

	var accommodationIds []primitive.ObjectID
	for id := range uniqueAccommodationIds {
		accommodationIds = append(accommodationIds, id)
	}

	listOfInvalidIds, err := rr.FindReservationForSearch(periodIDs, accommodationIds, dates.StartDate, dates.EndDate)
	if err != nil {
		rr.logger.Println(err)
		return ListOfObjectIds{}, err
	}

	return listOfInvalidIds, nil
}

func (rr *ReservationRepo) FindReservationForSearch(periodsIds []gocql.UUID, listOfAccommodationIds []primitive.ObjectID, startDate, endDate time.Time) (ListOfObjectIds, error) {
	idAccommodationsMap := make(map[primitive.ObjectID]Reservations)

	//need fixing
	//listOfAccommodationIds, _ := rr.FindAccommodationIdsByPeriods(periodsIds)

	for _, id := range listOfAccommodationIds {
		idAccommodationsMap[id] = Reservations{}
	}

	for _, id := range periodsIds {
		query := `SELECT id, id_accommodation, id_available_period, id_user, start_date, 
       end_date, guest_number, price 
       FROM reservations_by_available_period 
       WHERE id_available_period = ? `

		var (
			idAccommodationStr string
			idUserStr          string
			reservation        ReservationByAvailablePeriod
		)

		iter := rr.session.Query(query, id).Consistency(gocql.One).Iter()

		for iter.Scan(
			&reservation.ID, &idAccommodationStr, &reservation.IDAvailablePeriod, &idUserStr,
			&reservation.StartDate, &reservation.EndDate, &reservation.GuestNumber, &reservation.Price,
		) {
			reservation.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
			reservation.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

			newReservation := ReservationByAvailablePeriod{
				ID:                reservation.ID,
				IDAccommodation:   reservation.IDAccommodation,
				IDAvailablePeriod: reservation.IDAvailablePeriod,
				IDUser:            reservation.IDUser,
				StartDate:         reservation.StartDate,
				EndDate:           reservation.EndDate,
				GuestNumber:       reservation.GuestNumber,
				Price:             reservation.Price,
			}

			idAccommodation := newReservation.IDAccommodation
			idAccommodationsMap[idAccommodation] = append(idAccommodationsMap[idAccommodation], &newReservation)
		}

		if err := iter.Close(); err != nil {
			rr.logger.Println(err)
			return ListOfObjectIds{}, err
		}
	}

	idAccommodations := ListOfObjectIds{}

	for _, reservations := range idAccommodationsMap {
		for _, reservation := range reservations {
			if (startDate.After(reservation.StartDate) && startDate.Before(reservation.EndDate)) ||
				(endDate.After(reservation.StartDate) && endDate.Before(reservation.EndDate)) ||
				startDate.Equal(reservation.EndDate) || endDate.Equal(reservation.StartDate) ||
				endDate.Equal(reservation.EndDate) || startDate.Equal(reservation.StartDate) {
				delete(idAccommodationsMap, reservation.IDAccommodation)
				break
			}
		}
	}

	for id, _ := range idAccommodationsMap {
		idAccommodations.ObjectIds = append(idAccommodations.ObjectIds, id)
	}

	return idAccommodations, nil
}

func (rr *ReservationRepo) FindAccommodationIdsByPeriods(periodsIds []gocql.UUID) ([]primitive.ObjectID, error) {
	var accommodationIds []primitive.ObjectID

	for _, id := range periodsIds {
		iter := rr.session.Query(`
            SELECT id_accommodation 
            FROM available_periods_by_accommodation 
            WHERE id = ? ALLOW FILTERING`, id).Iter()

		var idAccommodationStr string
		for iter.Scan(&idAccommodationStr) {
			idAccommodation, err := primitive.ObjectIDFromHex(idAccommodationStr)
			if err != nil {
				rr.logger.Println(err)
				return nil, err
			}

			accommodationIds = append(accommodationIds, idAccommodation)
		}

		if err := iter.Close(); err != nil {
			rr.logger.Println(err)
			return nil, err
		}
	}

	return accommodationIds, nil
}

func (rr *ReservationRepo) FindReservationIdsByStartDate(dates *Dates) (ListOfObjectIds, error) {
	query := `
			SELECT id_accommodation 
			FROM reservations_by_available_period
			WHERE start_date <= ? AND end_date <= ?
			ALLOW FILTERING
			`

	iter := rr.session.Query(query, dates.StartDate, dates.StartDate).Iter()

	defer iter.Close()

	var listOfIds ListOfObjectIds

	var result string
	for iter.Scan(&result) {
		idAccommodation, err := primitive.ObjectIDFromHex(result)
		if err != nil {
			rr.logger.Println(err)
			return ListOfObjectIds{}, err
		}

		listOfIds.ObjectIds = append(listOfIds.ObjectIds, idAccommodation)
	}

	if err := iter.Close(); err != nil {
		rr.logger.Println(err)
		return ListOfObjectIds{}, errors.New("error closing iterator")
	}

	return listOfIds, nil
}

func (rr *ReservationRepo) FindReservationIdsByEndDate(dates *Dates) (ListOfObjectIds, error) {
	query := `
			SELECT id_accommodation 
			FROM reservations_by_available_period
			WHERE start_date <= ? AND end_date <= ?
			ALLOW FILTERING
			`

	iter := rr.session.Query(query, dates.EndDate, dates.EndDate).Iter()

	defer iter.Close()

	var listOfIds ListOfObjectIds

	var result string
	for iter.Scan(&result) {
		idAccommodation, err := primitive.ObjectIDFromHex(result)
		if err != nil {
			rr.logger.Println(err)
			return ListOfObjectIds{}, err
		}

		listOfIds.ObjectIds = append(listOfIds.ObjectIds, idAccommodation)
	}

	if err := iter.Close(); err != nil {
		rr.logger.Println(err)
		return ListOfObjectIds{}, errors.New("error closing iterator")
	}

	listOfReservationIds, err := rr.FindReservationIdsByStartDate(dates)
	if err != nil {
		rr.logger.Println(err)
		return ListOfObjectIds{}, err
	}

	// Create a map to efficiently check if an ID exists
	reservationIdMap := make(map[primitive.ObjectID]struct{})
	for _, id := range listOfReservationIds.ObjectIds {
		reservationIdMap[id] = struct{}{}
	}

	// Iterate over the existing list of IDs and add those that don't exist in the reservation map
	for _, id := range listOfIds.ObjectIds {
		if _, exists := reservationIdMap[id]; !exists {
			listOfReservationIds.ObjectIds = append(listOfReservationIds.ObjectIds, id)
		}
	}

	fmt.Println("OVERLAP IDS: ", listOfReservationIds.ObjectIds)

	return listOfReservationIds, nil
}

// **END OF SEARCH**

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
	// Convert ObjectID into hexadecimal string
	hexString := objectID.Hex()

	// Parse hexadecimal string into gocql.UUID
	uuid, err := gocql.ParseUUID(hexString)
	if err != nil {
		return gocql.UUID{}, err
	}

	return uuid, nil
}
