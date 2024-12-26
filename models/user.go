package models

import (
	"github.com/seebasoft/prompter/goback/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email"`
}

func (u User) DBName() string {
	return "core"
}

func (u User) CollectionName() string {
	return "users"
}

func (u User) ToBSON(addId bool) bson.M {
	if addId {
		return bson.M{
			"_id":      u.ID,
			"username": u.Username,
			"email":    u.Email,
		}
	} else {
		return bson.M{
			"username": u.Username,
			"email":    u.Email,
		}
	}
}

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
