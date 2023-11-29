package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Credentials struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	Password     string             `bson:"password" json:"password"`
	Email        string             `bson:"email" json:"email"`
	Role         string             `bson:"role" json:"role"`
	IsActivated  bool               `bson:"isActivated" json:"isActivated"`
	RecoveryUUID string             `bson:"recoveryUUID" json:"recoveryUUID"`
}

type ActivatioModel struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ActivationUUID string             `bson:"activationUUID" json:"activationUUID"`
	Username       string             `bson:"username" json:"username"`
	Time           time.Time          `bson:"time" json:"time"`
	Confirmed      bool               `bson:"confirmed" json:"confirmed"`
}

type ChangePasswordRequest struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (c *Credentials) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(c)
}

func (c *Credentials) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(c)
}

func (c *ChangePasswordRequest) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(c)
}

func (c *ActivatioModel) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(c)
}

func (c *ChangePasswordRequest) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(c)
}

func (c *ActivatioModel) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(c)
}
