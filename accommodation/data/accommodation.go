package data

import (
	"encoding/json"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Accommodation struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	Location  string             `json:"location" bson:"location"`
	Amenities []AmenityEnum      `json:"amenities" bson:"amenities"`
	MinGuests int                `json:"minGuests" bson:"minGuests"`
	MaxGuests int                `json:"maxGuests" bson:"maxGuests"`
}

type Dates struct {
	AccommodationIds []primitive.ObjectID
	StartDate        time.Time `json:"startDate"`
	EndDate          time.Time `json:"endDate"`
}

type ListOfObjectIds struct {
	ObjectIds []primitive.ObjectID `json:"objectIds"`
}

type AmenityEnum int

const (
	Essentials      AmenityEnum = iota //0
	WiFi                               //1
	Parking                            //2
	AirConditioning                    //3
	Kitchen                            //4
	TV                                 //5
	Pool                               //6
	PetFriendly                        //7
	HairDryer                          //8
	Iron                               //9
	IndoorFireplace                    //10
	Heating                            //11
	Washer                             //12
	Hangers                            //13
	HotWater                           //14
	PrivateBathroom                    //15
	Gym                                //16
	SmokingAllowed                     //17
)

func (a *Accommodation) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(a)
}

func (a *Accommodation) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(a)
}

func (a *Accommodation) GetAmenitiesAsInt() []int {
	amenitiesAsInt := make([]int, len(a.Amenities))
	for i, amenity := range a.Amenities {
		amenitiesAsInt[i] = int(amenity)
	}
	return amenitiesAsInt
}

func (a *Accommodation) SetAmenitiesFromInt(amenitiesAsInt []int) {
	a.Amenities = make([]AmenityEnum, len(amenitiesAsInt))
	for i, val := range amenitiesAsInt {
		a.Amenities[i] = AmenityEnum(val)
	}
}
