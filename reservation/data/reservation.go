package data

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gocql/gocql"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AvailablePeriodByAccommodation struct {
	ID              gocql.UUID
	IDAccommodation primitive.ObjectID // Partition key
	StartDate       time.Time          // Sort key
	EndDate         time.Time
	Price           float64
	PricePerGuest   bool
}

type ReservationByAvailablePeriod struct {
	ID                gocql.UUID
	IDAccommodation   primitive.ObjectID
	IDAvailablePeriod gocql.UUID // Partition key
	IDUser            primitive.ObjectID
	StartDate         time.Time // Sort key
	EndDate           time.Time
	GuestNumber       int16
	Price             float64
}

type Dates struct {
	AccommodationIds []primitive.ObjectID `json:"accommodationIds"`
	StartDate        time.Time            `json:"startDate"`
	EndDate          time.Time            `json:"endDate"`
}

type ListOfObjectIds struct {
	ObjectIds []primitive.ObjectID `json:"objectIds"`
}

type AvailablePeriodsByAccommodation []*AvailablePeriodByAccommodation
type Reservations []*ReservationByAvailablePeriod

func (r *ReservationByAvailablePeriod) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *ReservationByAvailablePeriod) FromJSON(re io.Reader) error {
	d := json.NewDecoder(re)
	return d.Decode(r)
}

func (r *Dates) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *Dates) FromJSON(re io.Reader) error {
	d := json.NewDecoder(re)
	return d.Decode(r)
}

func (r *Reservations) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *Reservations) FromJSON(re io.Reader) error {
	d := json.NewDecoder(re)
	return d.Decode(r)
}

func (r *ListOfObjectIds) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *ListOfObjectIds) FromJSON(re io.Reader) error {
	d := json.NewDecoder(re)
	return d.Decode(r)
}

func (r *AvailablePeriodByAccommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *AvailablePeriodByAccommodation) FromJSON(re io.Reader) error {
	d := json.NewDecoder(re)
	return d.Decode(r)
}

func (r *AvailablePeriodsByAccommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *AvailablePeriodsByAccommodation) FromJSON(re io.Reader) error {
	d := json.NewDecoder(re)
	return d.Decode(r)
}
