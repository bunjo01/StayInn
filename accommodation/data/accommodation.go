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
	Amenities []AmenityEnum `json:"amenities"`
	MinGuests int      `json:"minGuests"`
	MaxGuests int      `json:"maxGuests"`
}

type AmenityEnum int

const (
    Essentials AmenityEnum = iota //0
	WiFi //1
    Parking //2
    AirConditioning //3
    Kitchen //4
    TV //5
    Pool //6
	PetFriendly //7
	HairDryer //8
	Iron //9
	IndoorFireplace //10
	Heating //11
	Washer //12
	Hangers //13
	HotWater //14
	PrivateBathroom //15
	Gym //16
	SmokingAllowed //17
)

func (a *Accommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(a)
}

func (a *Accommodation) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(a)
}

// func (a *Accommodation) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
// 	var amenities []int
// 	if err := gocql.Unmarshal(info, data, &amenities); err != nil {
// 		return err
// 	}

// 	a.Amenities = make([]AmenityEnum, len(amenities))
// 	for i, val := range amenities {
// 		a.Amenities[i] = AmenityEnum(val)
// 	}

// 	return nil
// }
