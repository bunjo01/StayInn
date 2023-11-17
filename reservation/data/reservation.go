package data

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"io"
	"time"
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

//type Reservation struct {
//	ID                                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
//	IDAccommodations                   primitive.ObjectID `bson:"idAccommodations" json:"idAccommodations"`
//	ReservedPeriods                    *[]ReservedPeriod  `bson:"reservedPeriods" json:"reservedPeriods"`
//	PricePerGuestConfiguration         PricePerGuestConfiguration
//	PricePerAccommodationConfiguration PricePerAccommodationConfiguration
//	PricePerGuest                      bool
//}

//type ReservedPeriod struct {
//	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
//	IDUser         primitive.ObjectID `bson:"idUser" json:"idUser"`
//	StartDate      time.Time          `bson:"startDate" json:"startDate"`
//	EndDate        time.Time          `bson:"endDate" json:"endDate"`
//	NumberOfGuests int                `bson:"numberOfGuests" json:"numberOfGuests"`
//	Price          float64            `bson:"price" json:"price"`
//}
//
//type PricePerGuestConfiguration struct {
//	StandardPrice      float64 `bson:"standardPrice" json:"standardPrice"`
//	SummerSeasonPrice  float64 `bson:"summerSeasonPrice" json:"summerSeasonPrice"`
//	WinterSeasonPrice  float64 `bson:"winterSeasonPrice" json:"winterSeasonPrice"`
//	WeekendSeasonPrice float64 `bson:"weekendSeasonPrice" json:"weekendSeasonPrice"`
//}
//
//type PricePerAccommodationConfiguration struct {
//	StandardPrice      float64 `bson:"standardPrice" json:"standardPrice"`
//	SummerSeasonPrice  float64 `bson:"summerSeasonPrice" json:"summerSeasonPrice"`
//	WinterSeasonPrice  float64 `bson:"winterSeasonPrice" json:"winterSeasonPrice"`
//	WeekendSeasonPrice float64 `bson:"weekendSeasonPrice" json:"weekendSeasonPrice"`
//}

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
