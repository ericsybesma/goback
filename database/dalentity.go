package database

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DalEntity interface {
	DBName() string
	CollectionName() string
	ToBSON(bool) bson.M
	FromBSON(bson.M) DalEntity
	GetID() primitive.ObjectID
	SetID(id primitive.ObjectID)
}
