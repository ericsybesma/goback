package database

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DalEntity interface {
	DBName() string
	CollectionName() string
	FromBSON(bson.M) DalEntity
	GetID() primitive.ObjectID
	SetID(id primitive.ObjectID)
}

type DalRepo interface {
	Create(update DalEntity) (primitive.ObjectID, error)
	ReadByID(id primitive.ObjectID) (DalEntity, error)
	ReadByFilter(filter bson.M, page int64, pageSize int64) ([]DalEntity, error)
	ReadBSON(filter bson.M, page int64, pageSize int64) ([]bson.M, error)
	UpdateByID(id primitive.ObjectID, update DalEntity) (int64, error)
	DeleteByID(id primitive.ObjectID) (int64, error)
}