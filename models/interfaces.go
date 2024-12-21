package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/bson"
)

type DbEntity interface {
    DBName() string
    CollectionName() string
    ToBSON(bool) bson.M
    FromBSON(bson.M) DbEntity
    GetID() primitive.ObjectID
    SetID(id primitive.ObjectID)
}