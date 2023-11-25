package data

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	Host  Role = "HOST"
	Guest Role = "GUEST"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

type NewUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"password"`
	FirstName string             `bson:"firstName" json:"firstName"`
	LastName  string             `bson:"lastName" json:"lastName"`
	Email     string             `bson:"email" json:"email"`
	Address   string             `bson:"address" json:"address"`
	Role      Role               `bson:"role" json:"role"`
}

func (nu *NewUser) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(nu)
}

func (nu *NewUser) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(nu)
}
