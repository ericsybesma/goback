package database

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
)

type DalRepo interface {
	Create(update DalEntity) (primitive.ObjectID, error)
	ReadByID(id primitive.ObjectID) (DalEntity, error)
	ReadByFilter(filter bson.M, page int64, pageSize int64) ([]DalEntity, error)
	UpdateByID(id primitive.ObjectID, update DalEntity) (int64, error)
	DeleteByID(id primitive.ObjectID) (int64, error)
}

type DalRpc interface {
	Create(update DalEntity) (primitive.ObjectID, error)
	ReadByID(id primitive.ObjectID) (DalEntity, error)
	ReadByFilter(filter bson.M, page int64, pageSize int64) error
	UpdateByID(id primitive.ObjectID, update DalEntity) (int64, error)
	UpdatePartialByFilter(filter bson.M, update bson.M) (int64, error)
	DeleteByID(id primitive.ObjectID) (int64, error)
	DeleteByFilter(filter bson.M) (int64, error)
}