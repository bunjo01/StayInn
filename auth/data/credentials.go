package data

import (
	"encoding/json"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Credentials struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username    string             `bson:"username" json:"username"`
	Password    string             `bson:"password" json:"password"`
	Email       string             `bson:"email" json:"email"`
	Role        Role               `bson:"role" json:"role"`
	IsActivated bool               `bson:"isActivated" json:"isActivated"`
}

type ActivatioModel struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ActivationUUID string             `bson:"activationUUID" json:"activationUUID"`
	Username       string             `bson:"username" json:"username"`
	Confirmed      bool               `bson:"confirmed" json:"confirmed"`
}

func (c *Credentials) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(c)
}

func (c *Credentials) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(c)
}

func (c *ActivatioModel) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(c)
}

func (c *ActivatioModel) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(c)
}
