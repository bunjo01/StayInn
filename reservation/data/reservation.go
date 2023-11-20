package data

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gocql/gocql"
)

type AvailablePeriodByAccommodation struct {
	ID              gocql.UUID
	IDAccommodation gocql.UUID // Partition key
	StartDate       time.Time  // Sort key
	EndDate         time.Time
	Price           float64
	PricePerGuest   bool
}

type ReservationByAvailablePeriod struct {
	ID                gocql.UUID
	IDAccommodation   gocql.UUID
	IDAvailablePeriod gocql.UUID // Partition key
	IDUser            gocql.UUID
	StartDate         time.Time // Sort key
	EndDate           time.Time
	GuestNumber       int16
	Price             float64
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

func (r *Reservations) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *Reservations) FromJSON(re io.Reader) error {
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
