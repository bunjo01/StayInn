package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Reservation struct {
	ID                                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	IDAccommodations                   primitive.ObjectID `bson:"idAccommodations" json:"idAccommodations"`
	ReservedPeriods                    *[]ReservedPeriod  `bson:"reservedPeriods" json:"reservedPeriods"`
	PricePerGuestConfiguration         PricePerGuestConfiguration
	PricePerAccommodationConfiguration PricePerAccommodationConfiguration
	PricePerGuest                      bool
}

type Reservations []*Reservation

type ReservedPeriod struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	IDUser         primitive.ObjectID `bson:"idUser" json:"idUser"`
	StartDate      time.Time          `bson:"startDate" json:"startDate"`
	EndDate        time.Time          `bson:"endDate" json:"endDate"`
	NumberOfGuests int                `bson:"numberOfGuests" json:"numberOfGuests"`
	Price          float64            `bson:"price" json:"price"`
}

type PricePerGuestConfiguration struct {
	StandardPrice      float64 `bson:"standardPrice" json:"standardPrice"`
	SummerSeasonPrice  float64 `bson:"summerSeasonPrice" json:"summerSeasonPrice"`
	WinterSeasonPrice  float64 `bson:"winterSeasonPrice" json:"winterSeasonPrice"`
	WeekendSeasonPrice float64 `bson:"weekendSeasonPrice" json:"weekendSeasonPrice"`
}

type PricePerAccommodationConfiguration struct {
	StandardPrice      float64 `bson:"standardPrice" json:"standardPrice"`
	SummerSeasonPrice  float64 `bson:"summerSeasonPrice" json:"summerSeasonPrice"`
	WinterSeasonPrice  float64 `bson:"winterSeasonPrice" json:"winterSeasonPrice"`
	WeekendSeasonPrice float64 `bson:"weekendSeasonPrice" json:"weekendSeasonPrice"`
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

func (r *ReservedPeriod) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (r *ReservedPeriod) FromJSON(re io.Reader) error {
	d := json.NewDecoder(re)
	return d.Decode(r)
}
