package models

import (
	"time"

	"github.com/seebasoft/prompter/goback/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email"`
	Birthdate time.Time		    `bson:"birthdate" json:"birthdate"`
}

func (u User) DBName() string {
	return "core"
}

func (u User) CollectionName() string {
	return "users"
}

// func (u User) ToBSON() ([]byte, error) {
//     bsonUser, err := bson.Marshal(u)
// 	if err != nil {
// 		return nil, fmt.Errorf("bson marshaling error: %w", err)
// 	}
// 	return bsonUser, err
// }

func (u User) FromBSON(bsonUser bson.M) database.DalEntity {
	return &User{
		ID:       bsonUser["_id"].(primitive.ObjectID),
		Username: bsonUser["username"].(string),
		Email:    bsonUser["email"].(string),
	}
}

func (u User) GetID() primitive.ObjectID {
	return u.ID
}

func (u *User) SetID(id primitive.ObjectID) {
	u.ID = id
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