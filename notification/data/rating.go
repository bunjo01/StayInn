package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RatingHost struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GuestID       primitive.ObjectID `bson:"idGuest" json:"idGuest"`
	HostID        primitive.ObjectID `bson:"idHost" json:"idHost"`
	GuestUsername string             `bson:"guestUsername" json:"guestUsername"`
	HostUsername  string             `bson:"hostUsername" json:"hostUsername"`
	Time          time.Time          `bson:"time" json:"time"`
	Rate          int                `bson:"rate" json:"rate"`
}

type RatingAccommodation struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GuestID         primitive.ObjectID `bson:"idGuest" json:"idGuest"`
	HostID          primitive.ObjectID `bson:"idHost" json:"idHost"`
	GuestUsername   string             `bson:"guestUsername" json:"guestUsername"`
	HostUsername    string             `bson:"hostUsername" json:"hostUsername"`
	IDAccommodation primitive.ObjectID `bson:"idAccommodation" json:"idAccommodation"`
	Time            time.Time          `bson:"time" json:"time"`
	Rate            int                `bson:"rate" json:"rate"`
}

type AverageRatingAccommodation struct {
	AccommodationID primitive.ObjectID `bson:"idAccommodation" json:"idAccommodation"`
	AverageRating   float64            `bson:"averageRating" json:"averageRating"`
}

type AverageRatingHost struct {
	Username      string  `bson:"username" json:"username"`
	AverageRating float64 `bson:"averageRating" json:"averageRating"`
}

func (rh *RatingHost) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(rh)
}

func (rh *RatingHost) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(rh)
}

func (ra *RatingAccommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(ra)
}

func (ra *RatingAccommodation) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(ra)
}
