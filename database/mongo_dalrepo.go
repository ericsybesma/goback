package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDalRepo struct {
	client     *mongo.Client
	entityType DalEntity
	ctx        context.Context
}

// Adapt NewMongoDalRepo to return the interface
func NewMongoDalRepo(client *mongo.Client, entityType DalEntity, ctx context.Context) DalRepo {
	return &MongoDalRepo{client: client, entityType: entityType, ctx: ctx}
}

func (r *MongoDalRepo) Create(entity DalEntity) (primitive.ObjectID, error) {
	collection := r.client.Database(r.entityType.DBName()).Collection(r.entityType.CollectionName())
	result, err := collection.InsertOne(r.ctx, entity)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("creating entity: %w", err)
	}

	insertedID := result.InsertedID.(primitive.ObjectID)
	return insertedID, nil
}

func (r *MongoDalRepo) ReadByID(id primitive.ObjectID) (DalEntity, error) {
	collection := r.client.Database(r.entityType.DBName()).Collection(r.entityType.CollectionName())
	filter := bson.M{"_id": id}
	var bsonEntity bson.M
	err := collection.FindOne(r.ctx, filter).Decode(&bsonEntity)
	if err != nil {
		return nil, fmt.Errorf("getting entity by ID: %w", err)
	}
	return r.entityType.FromBSON(bsonEntity), nil
}

func (r *MongoDalRepo) ReadByFilter(filter bson.M, page int64, pageSize int64) ([]DalEntity, error) {
	bsonEntities, _ := r.ReadBSON(filter, page, pageSize)
	entities := make([]DalEntity, 0, len(bsonEntities))
	for _, bsonEntity := range bsonEntities {
		entities = append(entities, r.entityType.FromBSON(bsonEntity))
	}
	return entities, nil
}

func (r *MongoDalRepo) ReadBSON(filter bson.M, page int64, pageSize int64) ([]bson.M, error) {
	collection := r.client.Database(r.entityType.DBName()).Collection(r.entityType.CollectionName())
	findOptions := options.Find().SetSkip((page - 1) * pageSize).SetLimit(pageSize)
	cursor, err := collection.Find(r.ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("reading entities by filter: %w", err)
	}
	defer cursor.Close(r.ctx)
	var bsonEntities []bson.M
	if err = cursor.All(r.ctx, &bsonEntities); err != nil {
		return nil, fmt.Errorf("failed to decode entities: %w", err)
	}
	return bsonEntities, nil
}

func (r *MongoDalRepo) UpdateByID(id primitive.ObjectID, update DalEntity) (int64, error) {
	collection := r.client.Database(r.entityType.DBName()).Collection(r.entityType.CollectionName())
	filter := bson.M{"_id": r.entityType.GetID()}
	updateResult, err := collection.ReplaceOne(r.ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("updating entity: %w", err)
	}
	return updateResult.ModifiedCount, nil
}

func (r *MongoDalRepo) DeleteByID(id primitive.ObjectID) (int64, error) {
	collection := r.client.Database(r.entityType.DBName()).Collection(r.entityType.CollectionName())
	filter := bson.M{"_id": id}
	deleteResult, err := collection.DeleteOne(r.ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("deleting entity: %w", err)
	}
	return deleteResult.DeletedCount, nil
}