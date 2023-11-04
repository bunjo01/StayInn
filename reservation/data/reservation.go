package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Reservation struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	//IDAccommodations    primitive.ObjectID   `bson:"idAccommodations" json:"idAccommodations"`
	//AvailabilityPeriods []AvailabilityPeriod `bson:"availabilityPeriods" json:"availabilityPeriods"`
	//ReservationDate          time.Time          `bson:"reservationDate" json:"reservationDate"`
}

type Reservations []*Reservation

type AvailabilityPeriod struct {
	IDUser             primitive.ObjectID `bson:"idUser" json:"idUser"`
	ID                 primitive.ObjectID `bson:"_id" json:"id"`
	StartDate          time.Time          `bson:"startDate" json:"startDate"`
	EndDate            time.Time          `bson:"endDate" json:"endDate"`
	PriceConfiguration PriceConfiguration `bson:"priceConfiguration" json:"priceConfiguration"`
	IsAvailable        bool               `bson:"isAvailable" json:"isAvailable"`
}

type PriceConfiguration struct {
	NumberOfReservedGuests int     `bson:"numberOfReservedGuests" json:"numberOfReservedGuests"`
	PricePerGuest          float64 `bson:"pricePerGuest" json:"pricePerGuest"`
	PricePerAccommodation  float64 `bson:"pricePerAccommodation" json:"pricePerAccommodation"`
	UsePricePerGuest       bool    `bson:"usePricePerGuest" json:"usePricePerGuest"`
}

func (r *Reservation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *Reservation) FromJSON(re io.Reader) error {
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
