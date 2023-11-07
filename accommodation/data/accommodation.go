package data

import (
	"encoding/json"
	"io"

	"github.com/gocql/gocql"
)

type Accommodation struct {
	ID        gocql.UUID   `json:"id"`
	Name      string   `json:"name"`
	Location  string   `json:"location"`
	Amenities []string `json:"amenities"`
	MinGuests int      `json:"min_guests"`
	MaxGuests int      `json:"max_guests"`
}

func (a *Accommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(a)
}

func (a *Accommodation) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(a)
}
