package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RatingHost struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GuestUsername string             `bson:"guestUsername" json:"guestUsername"`
	HostUsername  string             `bson:"hostUsername" json:"hostUsername"`
	Time          time.Time          `bson:"time" json:"time"`
}

type RatingAccommodation struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GuestUsername   string             `bson:"guestUsername" json:"guestUsername"`
	IDAccommodation primitive.ObjectID `bson:"idAccommodation" json:"idAccommodation"`
	Time            time.Time          `bson:"time" json:"time"`
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
