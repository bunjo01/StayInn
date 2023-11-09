package data

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
)

type AccommodationRepository struct {
	session *gocql.Session
	logger *log.Logger
}

func NewAccommodationRepository(logger *log.Logger) (*AccommodationRepository, error) {
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
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, "accommodation", 1)).Exec()
    if err != nil {
        logger.Println(err)
    }
	session.Close()

	cluster.Keyspace = "accommodation"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	return &AccommodationRepository{
		session: session,
		logger:  logger,
	}, nil
}

func (ar *AccommodationRepository) CloseSession() {
	ar.session.Close()
}

func (a *AccommodationRepository) CreateAccommodationTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS accommodations (
            id UUID PRIMARY KEY,
            name TEXT,
            location TEXT,
            amenities SET<INT>,
            min_guests INT,
            max_guests INT
        )
    `
    return a.session.Query(query).Exec()
}


func (ar *AccommodationRepository) CreateAccommodation(ctx context.Context, accommodation *Accommodation) error {
	amenitiesAsInt := make([]int, len(accommodation.Amenities))
	for i, amenity := range accommodation.Amenities {
		amenitiesAsInt[i] = int(amenity)
	}
	query := ar.session.Query(
		"INSERT INTO accommodations (id, name, location, amenities, min_guests, max_guests) VALUES (?, ?, ?, ?, ?, ?)",
		accommodation.ID, accommodation.Name, accommodation.Location, amenitiesAsInt, accommodation.MinGuests, accommodation.MaxGuests,
	)
	if err := query.Exec(); err != nil {
		ar.logger.Fatal(err.Error())
		return err
	}
	return nil
}

func (ar *AccommodationRepository) GetAllAccommodations(ctx context.Context) ([]*Accommodation, error) {
    query := "SELECT * FROM accommodations"
    iter := ar.session.Query(query).Iter()

    var accommodations []*Accommodation

    for {
        accommodation := &Accommodation{}
        var amenities CustomSetInt

        if !iter.Scan(&accommodation.ID, &accommodation.Name, &accommodation.Location, &amenities, &accommodation.MinGuests, &accommodation.MaxGuests) {
            break
        }

        accommodation.Amenities = make([]AmenityEnum, len(amenities.Values))
        for i, val := range amenities.Values {
            accommodation.Amenities[i] = AmenityEnum(val)
        }

        accommodations = append(accommodations, accommodation)
    }

    if err := iter.Close(); err != nil {
        ar.logger.Fatal(err.Error())
        return nil, err
    }

    return accommodations, nil
}



func (ar *AccommodationRepository) GetAccommodation(ctx context.Context, id gocql.UUID) (*Accommodation, error) {
	var accommodation Accommodation

	query := ar.session.Query(
		"SELECT id, name, location, amenities, min_guests, max_guests FROM accommodations WHERE id = ?",
		id,
	)

	if err := query.Scan(&accommodation.ID, &accommodation.Name, &accommodation.Location, &accommodation.Amenities, &accommodation.MinGuests, &accommodation.MaxGuests); err != nil {
		ar.logger.Fatal(err.Error())
		return nil, err
	}

	return &accommodation, nil
}


func (ar *AccommodationRepository) UpdateAccommodation(ctx context.Context, accommodation *Accommodation) error {
	// Implementacija za ažuriranje smeštaja u Cassandra bazi
	return nil
}

func (ar *AccommodationRepository) DeleteAccommodation(ctx context.Context, id gocql.UUID) error {
	// Implementacija za brisanje smeštaja iz Cassandra baze
	return nil
}
