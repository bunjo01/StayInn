package data

import (
	"encoding/json"
	"io"
)

type ActivationRequest struct {
	ActivationUUID string `bson:"activationUUID" json:"activationUUID"`
	Email          string `bson:"email" json:"email"`
}

func (c *ActivationRequest) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(c)
}

func (c *ActivationRequest) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(c)
}
