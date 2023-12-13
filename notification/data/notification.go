package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	HostID       primitive.ObjectID `bson:"hostID" json:"hostID"`
	HostUsername string             `bson:"hostUsername" json:"hostUsername"`
	Text         string             `bson:"text" json:"text"`
	Time         time.Time          `bson:"time" json:"time"`
}

func (n *Notification) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(n)
}

func (n *Notification) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(n)
}
