package models

import (
	"time"

	"github.com/seebasoft/prompter/goback/dal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"errors"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Birthdate time.Time          `bson:"birthdate" json:"birthdate"`
}

func (u *User) Namespace() string {
	return "core"
}

func (u *User) ItemGroup() string {
	return "users"
}

func (u *User) Marshal() ([]byte, error) {
	return bson.Marshal(u)
}

func (u *User) Unmarshal(raw []byte) error {
	if err := bson.Unmarshal(raw, u); err != nil {
		return err
	}

	if u.Birthdate.IsZero() {
		u.Birthdate = time.Time{}
	}

	if u.ID == primitive.NilObjectID {
		u.ID = primitive.NewObjectID()
	}

	if u.Username == "" {
		u.Username = ""
	}

	if u.Email == "" {
		u.Email = ""
	}

	return nil
}


func (u *User) New() dal.Item {
	return &User{}
}	

func (u *User) GetKey() interface{} {
	return u.ID
}	

var ErrInvalidKey = errors.New("invalid key")

func (u *User) SetKey(key interface{}) error {
	if id, ok := key.(primitive.ObjectID); ok {
		u.ID = id
		return nil
	}
	return ErrInvalidKey
}

// // User represents a user in the system.
// type User struct {
// 	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // MongoDB ID
// 	FirstName string             `bson:"first_name" json:"firstName" validate:"required,max=50"`
// 	LastName  string             `bson:"last_name" json:"lastName" validate:"required,max=50"`
// 	Username  string             `bson:"username" json:"username" validate:"required,min=3,max=20,alphanum"` // alphanumeric
// 	Email     string             `bson:"email" json:"email" validate:"required,email"`
// 	Password  string             `bson:"password,omitempty" json:"password,omitempty" validate:"omitempty,min=8"` // omitempty for responses, min length for creation/updates
// 	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
// 	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
// 	// Optional fields
// 	Birthdate *time.Time         `bson:"birthdate,omitempty" json:"birthdate,omitempty"`
// 	Phone     *string            `bson:"phone,omitempty" json:"phone,omitempty"`
// 	Address   *Address           `bson:"address,omitempty" json:"address,omitempty"`
// 	Roles     []string           `bson:"roles,omitempty" json:"roles,omitempty"` // User roles/permissions
// 	Active    bool               `bson:"active" json:"active"`                 // User status
// }

// // Address represents a user's address.
// type Address struct {
// 	Street     string  `bson:"street" json:"street"`
// 	City       string  `bson:"city" json:"city"`
// 	State      string  `bson:"state" json:"state"`
// 	PostalCode string  `bson:"postalCode" json:"postalCode"`
// 	Country    string  `bson:"country" json:"country"`
// }
