package data

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gocql/gocql"
)

type ReservationRepo struct {
	session *gocql.Session
}

// Constructor
func New(session *gocql.Session) (*ReservationRepo, error) {
	return &ReservationRepo{
		session: session,
	}, nil
}

// Disconnect
func (rr *ReservationRepo) CloseSession() {
	rr.session.Close()
}

func (rr *ReservationRepo) CreateTables() error {
	err := rr.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s 
        (id UUID, id_accommodation TEXT, id_user TEXT, start_date TIMESTAMP, end_date TIMESTAMP, 
        price DOUBLE, price_per_guest BOOLEAN, 
        PRIMARY KEY ((id_accommodation), id)) 
        WITH CLUSTERING ORDER BY (id DESC)`,
			"available_periods_by_accommodation")).Exec()
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#1 Error while creating database tables: %v", err))
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
		log.Fatal(fmt.Sprintf("[rese-repo]rr#2 Error while creating database tables: %v", err))
		return err
	}

	return nil
}

func (rr *ReservationRepo) GetAvailablePeriodsByAccommodation(id string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`
		SELECT id, id_accommodation, id_user, start_date, end_date, price, price_per_guest 
		FROM available_periods_by_accommodation WHERE id_accommodation = ?`,
		id).Iter().Scanner()

	var availablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var period AvailablePeriodByAccommodation
		var idAccommodationStr string
		var idUserStr string

		err := scanner.Scan(&period.ID, &idAccommodationStr, &idUserStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)
		if err != nil {
			log.Fatal(fmt.Sprintf("[rese-repo]rr#3 Error while scanning from database: %v", err))
			return nil, err
		}

		idAccommodation, err := primitive.ObjectIDFromHex(idAccommodationStr)
		if err != nil {
			log.Error(fmt.Sprintf("[rese-repo]rr#4 Error while parsing id: %v", err))
			return nil, err
		}
		period.IDAccommodation = idAccommodation

		idUser, err := primitive.ObjectIDFromHex(idUserStr)
		if err != nil {
			log.Error(fmt.Sprintf("[rese-repo]rr#5 Error while parsing id: %v", err))
			return nil, err
		}
		period.IDUser = idUser

		availablePeriods = append(availablePeriods, &period)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#6 Error while scanning from database: %v", err))
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
			log.Fatal(fmt.Sprintf("[rese-repo]rr#7 Error while scanning from database: %v", err))
			return nil, err
		}

		// Convert idAccommodationStr and idUserStr strings to primitive.ObjectID
		idAccommodation, err := primitive.ObjectIDFromHex(idAccommodationStr)
		if err != nil {
			log.Error(fmt.Sprintf("[rese-repo]rr#8 Error while parsing id: %v", err))
			return nil, err
		}
		reservation.IDAccommodation = idAccommodation

		idUser, err := primitive.ObjectIDFromHex(idUserStr)
		if err != nil {
			log.Error(fmt.Sprintf("[rese-repo]rr#9 Error while parsing id: %v", err))
			return nil, err
		}
		reservation.IDUser = idUser

		reservations = append(reservations, &reservation)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#10 Error while scanning from database: %v", err))
		return nil, err
	}

	return reservations, nil
}

// extract username from token and communicate with profile service
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
		log.Error(fmt.Sprintf("[rese-repo]rr#11 Error while checking overlap of dates: %v", err))
		return err
	}

	if isOverLap {
		err = errors.New("date overlap")
		return err
	}

	availablePeriodId, _ := gocql.RandomUUID()
	idAccommodation := availablePeriod.IDAccommodation.Hex()
	idUser := availablePeriod.IDUser.Hex()
	err = rr.session.Query(
		`INSERT INTO available_periods_by_accommodation (id, id_accommodation, id_user, start_date, end_date, price, price_per_guest) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		availablePeriodId, idAccommodation, idUser, availablePeriod.StartDate, availablePeriod.EndDate,
		availablePeriod.Price, availablePeriod.PricePerGuest).Exec()
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#12 Error while inserting in database: %v", err))
		return err
	}

	return nil
}

func (rr *ReservationRepo) InsertReservationByAvailablePeriod(reservation *ReservationByAvailablePeriod) error {
	reservationId, _ := gocql.RandomUUID()

	// Check if the reservation is within the appropriate range of the available period
	availablePeriod, err := rr.FindAvailablePeriodById(reservation.IDAvailablePeriod.String(), reservation.IDAccommodation.Hex())
	if err != nil {
		log.Error(fmt.Sprintf("[rese-repo]rr#13 Error while finding available period by id: %v", err))
		return err
	}
	if reservation.StartDate.Before(availablePeriod.StartDate) || reservation.EndDate.After(availablePeriod.EndDate) {
		log.Error(fmt.Sprintf("[rese-repo]rr#14 Error while comparing two dates: %v", err))
		return errors.New("reservation is not within the appropriate range of the available period")
	}

	if reservation.EndDate.Sub(reservation.StartDate) < 24*time.Hour {
		log.Error(fmt.Sprintf("[rese-repo]rr#15 Error while creating reservation: %v", err))
		return errors.New("EndDate must be at least one day after StartDate")
	}

	// Retrieve existing reservations for the available period
	existingReservations, err := rr.FindAllReservationsByAvailablePeriod(availablePeriod.ID.String())
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#16 Error while getting all available periods: %v", err))
		return err
	}

	// Check for overlapping reservations
	for _, existingReservation := range existingReservations {
		if (reservation.StartDate.Before(existingReservation.EndDate) || reservation.StartDate.Equal(existingReservation.EndDate)) &&
			(reservation.EndDate.After(existingReservation.StartDate) || reservation.EndDate.Equal(existingReservation.StartDate)) {
			log.Error(fmt.Sprintf("[rese-repo]rr#17 Error while checking for reservation overlap: %v", err))
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
		log.Fatal(fmt.Sprintf("[rese-repo]rr#18 Error while inserting in database: %v", err))
		return err
	}

	return nil
}

// Add so only user who make period can update it, extract username from token and communicate with profile service
func (rr *ReservationRepo) UpdateAvailablePeriodByAccommodation(availablePeriod *AvailablePeriodByAccommodation) error {
	id := availablePeriod.ID
	accommodationdId := availablePeriod.IDAccommodation.Hex()

	availablePeriods, err := rr.FindAvailablePeriodsById(id.String(), accommodationdId)
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#19 Error while finding available period by id: %v", err))
		return err
	}

	if len(availablePeriods) != 1 {
		log.Error(fmt.Sprintf("[rese-repo]rr#20 Error while finding available period by id: %v", err))
		return err
	}

	reservations, err := rr.GetReservationsByAvailablePeriod(id.String())
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#21 Error while finding reservation by available period: %v", err))
		return err
	}

	if len(reservations) != 0 {
		log.Error(fmt.Sprintf("[rese-repo]rr#22 Error while chaning period with reservations: %v", err))
		err = errors.New("cannot change period with reservations")
		return err
	}

	isOverLap, err := rr.checkForOverlap(*availablePeriod, accommodationdId)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-repo]rr#23 Error while checking for period overlap: %v", err))
		return err
	}

	if isOverLap {
		log.Error(fmt.Sprintf("[rese-repo]rr#24 Error while checking for period overlap: %v", err))
		err = errors.New("date overlap")
		return err
	}

	if availablePeriod.Price < 0 {
		log.Error(fmt.Sprintf("[rese-repo]rr#25 Error while creating period: %v", err))
		err = errors.New("price cannot be negative")
		return err
	}

	if availablePeriod.StartDate.Before(time.Now()) {
		log.Error(fmt.Sprintf("[rese-repo]rr#26 Error while creating period: %v", err))
		err = errors.New("start date must be in the future")
		return err
	}

	if availablePeriod.StartDate.After(availablePeriod.EndDate) {
		log.Error(fmt.Sprintf("[rese-repo]rr#27 Error while creating period: %v", err))
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
		log.Fatal(fmt.Sprintf("[rese-repo]rr#28 Error while inserting in database: %v", err))
		return err
	}

	return nil
}

func (rr *ReservationRepo) FindAvailablePeriodsByAccommodationId(accommodationId string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`
    SELECT id, id_accommodation, id_user, start_date, end_date, price, price_per_guest 
    FROM available_periods_by_accommodation 
    WHERE id_accommodation = ?`, accommodationId).Iter().Scanner()

	var availablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var (
			idAccommodationStr string
			idUserStr          string
			period             AvailablePeriodByAccommodation
		)

		err := scanner.Scan(&period.ID, &idAccommodationStr, &idUserStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)

		if err != nil {
			log.Fatal(fmt.Sprintf("[rese-repo]rr#29 Error while scanning from database: %v", err))
			return nil, err
		}

		// Convert string to primitive.ObjectID
		period.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
		period.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

		availablePeriods = append(availablePeriods, &period)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#30 Error while scanning from database: %v", err))
		return nil, err
	}

	return availablePeriods, nil
}

func (rr *ReservationRepo) FindAvailablePeriodsById(id, accommodationId string) (AvailablePeriodsByAccommodation, error) {
	scanner := rr.session.Query(`
        SELECT id, id_accommodation, id_user, start_date, end_date, price, price_per_guest 
        FROM available_periods_by_accommodation 
        WHERE id = ? AND id_accommodation = ?`,
		id, accommodationId).Iter().Scanner()

	var availablePeriods AvailablePeriodsByAccommodation
	for scanner.Next() {
		var (
			idAccommodationStr string
			idUserStr          string
			period             AvailablePeriodByAccommodation
		)

		err := scanner.Scan(&period.ID, &idAccommodationStr, &idUserStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)

		if err != nil {
			log.Fatal(fmt.Sprintf("[rese-repo]rr#31 Error while scanning from database: %v", err))
			return nil, err
		}

		period.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
		period.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

		availablePeriods = append(availablePeriods, &period)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#32 Error while scanning from database: %v", err))
		return nil, err
	}

	return availablePeriods, nil
}

func (rr *ReservationRepo) FindAvailablePeriodById(id, accommodationID string) (*AvailablePeriodByAccommodation, error) {
	query := `SELECT id, id_accommodation, id_user, start_date, end_date, price, price_per_guest 
          FROM available_periods_by_accommodation 
          WHERE id = ? AND id_accommodation = ? LIMIT 1`

	var (
		idAccommodationStr string
		idUserStr          string
		period             AvailablePeriodByAccommodation
	)

	err := rr.session.Query(query, id, accommodationID).Consistency(gocql.One).Scan(
		&period.ID, &idAccommodationStr, &idUserStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest,
	)

	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#33 Error while scanning from database: %v", err))
		return nil, err
	}

	// Convert string to primitive.ObjectID
	period.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
	period.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

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
			log.Fatal(fmt.Sprintf("[rese-repo]rr#34 Error while scanning from database: %v", err))
			return nil, err
		}

		// Convert strings to primitive.ObjectID
		reservation.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
		reservation.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

		reservations = append(reservations, &reservation)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#35 Error while scanning from database: %v", err))
		return nil, err
	}

	return reservations, nil
}

func (rr *ReservationRepo) FindAllReservationsByUserID(userID string) (Reservations, error) {
	scanner := rr.session.Query(`
        SELECT id, id_accommodation, id_available_period, id_user, start_date, end_date, guest_number, price
        FROM reservations_by_available_period
        WHERE id_user = ? ALLOW FILTERING`, userID).Iter().Scanner()

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
			log.Fatal(fmt.Sprintf("[rese-repo]rr#36 Error while scanning from database: %v", err))
			return nil, err
		}

		// Convert strings to primitive.ObjectID
		reservation.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
		reservation.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

		reservations = append(reservations, &reservation)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#37 Error while scanning from database: %v", err))
		return nil, err
	}

	return reservations, nil
}

func (rr *ReservationRepo) FindAllReservationsByUserIDExpired(userID string) (Reservations, error) {
	scanner := rr.session.Query(`
        SELECT id, id_accommodation, id_available_period, id_user, start_date, end_date, guest_number, price
        FROM reservations_by_available_period
        WHERE id_user = ? AND end_date < ?
        ALLOW FILTERING`, userID, time.Now()).Iter().Scanner()

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
			log.Fatal(fmt.Sprintf("[rese-repo]rr#38 Error while scanning from database: %v", err))
			return nil, err
		}

		// Convert strings to primitive.ObjectID
		reservation.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
		reservation.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

		reservations = append(reservations, &reservation)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#39 Error while scanning from database: %v", err))
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
		log.Fatal(fmt.Sprintf("[rese-repo]rr#40 Error while scanning from database: %v", err))
		return nil, err
	}

	// Convert strings to primitive.ObjectID
	reservation.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
	reservation.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

	return &reservation, nil
}

func (rr *ReservationRepo) DeleteReservationByIdAndAvailablePeriodID(id, periodID, ownerId string) error {
	reservation, err := rr.FindReservationByIdAndAvailablePeriod(id, periodID)
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#41 Error while finging reservation by id and period: %v", err))
		return err
	}

	if reservation.IDUser.Hex() != ownerId {
		log.Error(fmt.Sprintf("[rese-repo]rr#42 Error while comparing userid and ownerid: %v", err))
		return errors.New("you are not owner of reservation")
	}

	if time.Now().After(reservation.StartDate) {
		log.Error(fmt.Sprintf("[rese-repo]rr#43 Error while comparing present with reservation start date: %v", err))
		return errors.New("cannot delete reservation after start date has passed")
	}

	query := `DELETE FROM reservations_by_available_period
              WHERE id = ? AND id_available_period = ?`

	if err := rr.session.Query(query, id, periodID).Exec(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#44 Error while retriving data from database: %v", err))
		return err
	}

	return nil
}

func (rr *ReservationRepo) CheckAndDeleteReservationsByUserID(userID primitive.ObjectID) error {
	reservations, err := rr.FindAllReservationsByUserID(userID.Hex())
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#45 Error while finding reservations by userid: %v", err))
		return err
	}

	processedAccommodations := make(map[primitive.ObjectID]bool)
	// Check if any reservation has an end date in the future
	for _, reservation := range reservations {
		if time.Now().Before(reservation.EndDate) {
			log.Error(fmt.Sprintf("[rese-repo]rr#46 Error while finding user active reservation: %v", err))
			return errors.New("user has active reservations")
		}
		// Mark the accommodation as processed
		processedAccommodations[reservation.IDAccommodation] = true
	}

	for accommodationID := range processedAccommodations {
		query := `DELETE FROM reservations_by_available_period
              WHERE id_accommodation = ? AND id_user = ?`

		if err := rr.session.Query(query, accommodationID.Hex(), userID.Hex()).Exec(); err != nil {
			log.Fatal(fmt.Sprintf("[rese-repo]rr#47 Error while retriving data from database: %v", err))
			return err
		}
	}

	return nil
}

func (rr *ReservationRepo) DeletePeriodsForAccommodations(accIDs []primitive.ObjectID) error {
	for _, accID := range accIDs {
		periods, err := rr.FindAvailablePeriodsByAccommodationId(accID.String())
		if err != nil {
			log.Fatal(fmt.Sprintf("[rese-repo]rr#48 Error while finding periods by accommodation id: %v", err))
			return err
		}

		for _, period := range periods {
			reservations, err := rr.FindAllReservationsByAvailablePeriod(period.ID.String())
			if err != nil {
				log.Error(fmt.Sprintf("[rese-repo]rr#49 Error while finding reservations by period: %v", err))
				return err
			}

			var reservationIDs []gocql.UUID
			for _, reservation := range reservations {
				if !time.Now().After(reservation.EndDate) {
					// If the end date has not passed, disallow deletion and return an error
					log.Error(fmt.Sprintf("[rese-repo]rr#50 Error while deleting period with active reservations: %v", err))
					return errors.New("cannot delete period, there are active reservations")
				}

				reservationIDs = append(reservationIDs, reservation.ID)
			}

			// Batch deletion of reservations
			if len(reservationIDs) > 0 {
				query := `DELETE FROM reservations_by_available_period WHERE id IN ?`

				if err := rr.session.Query(query, reservationIDs).Exec(); err != nil {
					log.Error(fmt.Sprintf("[rese-repo]rr#51 Error while deleting data from databse: %v", err))
					return err
				}
			}

			query := `DELETE FROM available_periods_by_accommodation WHERE id = ?`

			if err := rr.session.Query(query, period.ID).Exec(); err != nil {
				log.Fatal(fmt.Sprintf("[rese-repo]rr#52 Error while deleting data from databse: %v", err))
				return err
			}
		}
	}

	return nil
}

func (rr *ReservationRepo) FindAccommodationIdsByDates(dates *Dates) (ListOfObjectIds, error) {
	var periodIDs []gocql.UUID
	var accommodationIds []primitive.ObjectID
	uniqueAccommodationIds := make(map[primitive.ObjectID]struct{})

	for _, id := range dates.AccommodationIds {
		scanner := rr.session.Query(`
			SELECT id, id_accommodation, id_user, start_date, end_date, price, price_per_guest 
			FROM available_periods_by_accommodation 
			WHERE id_accommodation = ?`, id.Hex()).Iter().Scanner()

		for scanner.Next() {
			var (
				idAccommodationStr string
				idUserStr          string
				period             AvailablePeriodByAccommodation
			)

			err := scanner.Scan(&period.ID, &idAccommodationStr, &idUserStr, &period.StartDate, &period.EndDate, &period.Price, &period.PricePerGuest)

			if err != nil {
				log.Fatal(fmt.Sprintf("[rese-repo]rr#53 Error while scanning from databse: %v", err))
				return ListOfObjectIds{}, err
			}

			period.IDAccommodation, _ = primitive.ObjectIDFromHex(idAccommodationStr)
			period.IDUser, _ = primitive.ObjectIDFromHex(idUserStr)

			if (period.StartDate.Before(dates.StartDate) || period.StartDate.Equal(dates.StartDate)) &&
				(period.EndDate.After(dates.EndDate) || period.EndDate.Equal(dates.EndDate)) {
				periodIDs = append(periodIDs, period.ID)
				uniqueAccommodationIds[period.IDAccommodation] = struct{}{}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(fmt.Sprintf("[rese-repo]rr#54 Error while scanning data from databse: %v", err))
			return ListOfObjectIds{}, err
		}
	}

	for id := range uniqueAccommodationIds {
		accommodationIds = append(accommodationIds, id)
	}

	listOfInvalidIds, err := rr.FindReservationForSearch(periodIDs, accommodationIds, dates.StartDate, dates.EndDate)
	if err != nil {
		log.Error(fmt.Sprintf("[rese-repo]rr#55 Error while finding reservation for search: %v", err))
		return ListOfObjectIds{}, err
	}

	return listOfInvalidIds, nil
}

func (rr *ReservationRepo) FindReservationForSearch(periodsIds []gocql.UUID, listOfAccommodationIds []primitive.ObjectID, startDate, endDate time.Time) (ListOfObjectIds, error) {
	idAccommodationsMap := make(map[primitive.ObjectID]Reservations)

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
			log.Error(fmt.Sprintf("[rese-repo]rr#56 Error while itering over objects: %v", err))
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

func (rr *ReservationRepo) GetDistinctIds(idColumnName string, tableName string) ([]string, error) {
	scanner := rr.session.Query(
		fmt.Sprintf(`SELECT DISTINCT %s FROM %s`, idColumnName, tableName)).Iter().Scanner()
	var ids []string
	for scanner.Next() {
		var id string
		err := scanner.Scan(&id)
		if err != nil {
			log.Fatal(fmt.Sprintf("[rese-repo]rr#57 Error while scanning for ids: %v", err))
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#58 Error while scanning data from databse: %v", err))
		return nil, err
	}
	return ids, nil
}

func (rr *ReservationRepo) checkForOverlap(newPeriod AvailablePeriodByAccommodation, accommodationId string) (bool, error) {
	avalablePeriods, err := rr.FindAvailablePeriodsByAccommodationId(accommodationId)
	if err != nil {
		log.Fatal(fmt.Sprintf("[rese-repo]rr#59 Error while finding available periods by accommodation: %v", err))
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
