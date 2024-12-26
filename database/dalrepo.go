package database

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
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
	UpdateByID(id primitive.ObjectID, update DalEntity) (int64, error)
	DeleteByID(id primitive.ObjectID) (int64, error)
}

type DalRpc interface {
	DalRepo
	UpdatePartialByFilter(filter bson.M, update bson.M) (int64, error)
	DeleteByFilter(filter bson.M) (int64, error)
}