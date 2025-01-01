package database

import (
	"context"
	"fmt"
	//"regexp"

	"github.com/seebasoft/prompter/goback/dal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoKey implements dal.Key for MongoDB.
type MongoKey dal.AnyKey

func NewMongoKey(id primitive.ObjectID) dal.Key {
	return dal.NewAnyKey(id)
}

// MongoFilter implements dal.Filter for MongoDB.
type MongoFilter struct {
	nativeFilter bson.M
}

func NewMongoFilter(filter bson.M) dal.Filter {
	return &MongoFilter{nativeFilter: filter}
}

func (f *MongoFilter) ToNative() interface{} {
	return f.nativeFilter
}

// MongoDalQueryOptions
type MongoDalQueryOptions struct {
	filter dal.Filter
	sort   interface{}
	limit  int64
	skip   int64
}

func (m MongoDalQueryOptions) GetFilter() dal.Filter {
	return m.filter
}

func (m MongoDalQueryOptions) GetSort() interface{} {
	return m.sort
}

func (m MongoDalQueryOptions) GetLimit() int64 {
	return m.limit
}

func (m MongoDalQueryOptions) GetSkip() int64 {
	return m.skip
}

func NewMongoDalQueryOptions(filter dal.Filter, sort interface{}, limit int64, skip int64) dal.QueryOptions {
	return &MongoDalQueryOptions{
		filter: filter,
		sort:   sort,
		limit:  limit,
		skip:   skip,
	}
}

type mongoItemIterator[T any] struct {
	cursor *mongo.Cursor
	err    error
}

func (m *mongoItemIterator[T]) Next(ctx context.Context) bool {
	if m.err != nil {
		return false // Return false if there's a previous error
	}
	return m.cursor.Next(ctx)
}

func (m *mongoItemIterator[T]) Decode(item T) error {
	if m.err != nil {
		return m.err
	}
	m.err = m.cursor.Decode(item)
	return m.err
}

func (m *mongoItemIterator[T]) Close(ctx context.Context) error {
	if m.cursor != nil {
		return m.cursor.Close(ctx)
	}
	return nil
}

func (m *mongoItemIterator[T]) Err() error {
	return m.err
}

type MongoStore struct {
	client *mongo.Client
}

// Adapt NewMongoStore to return the interface
func NewMongoStore(client *mongo.Client) dal.Store {
	return &MongoStore{client: client}
}

func (r *MongoStore) Create(ctx context.Context, item dal.Item) (dal.Item, error) {
	collection := r.client.Database(item.Namespace()).Collection(item.ItemGroup())
	result, err := collection.InsertOne(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("creating entity: %w", err)
	}

	insertedID := result.InsertedID.(primitive.ObjectID)
	item.SetKey(insertedID)
	return item, nil
}

func (r *MongoStore) ReadByKey(ctx context.Context, key interface{}, item dal.Item) error {
	collection := r.client.Database(item.Namespace()).Collection(item.ItemGroup())
	filter := bson.M{"_id": key}
	var raw bson.Raw

	err := collection.FindOne(ctx, filter).Decode(&raw)
	if err != nil {
		return fmt.Errorf("getting entity by ID: %w", err)
	}

	item.Unmarshal(raw)
	return nil
}

func (r *MongoStore) ReadByFilter(ctx context.Context, opts dal.QueryOptions, itemType dal.Item) (dal.ItemIterator, error) {
	findOptions := options.Find()
	
	// Check if sort is not nil AND not empty
	if sortMap, ok := opts.GetSort().(bson.D); ok && len(sortMap) > 0 { 
		findOptions.SetSort(sortMap)
	}
	if opts.GetLimit() > 0 {
		findOptions.SetLimit(opts.GetLimit())
	}
	if opts.GetSkip() > 0 {
		findOptions.SetSkip(opts.GetSkip())
	}

	collection := r.client.Database(itemType.Namespace()).Collection(itemType.ItemGroup())
	cursor, err := collection.Find(ctx, opts.GetFilter().ToNative(), findOptions)
	if err != nil {
		return nil, fmt.Errorf("finding by filter: %w", err)
	}

	return &mongoItemIterator[dal.Item]{cursor: cursor}, nil
}

func (r *MongoStore) UpdateByKey(ctx context.Context, key interface{}, update dal.Item) (int64, error) {
	collection := r.client.Database(update.Namespace()).Collection(update.ItemGroup())
	filter := bson.M{"_id": key}
	updateResult, err := collection.ReplaceOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("updating entity: %w", err)
	}
	return updateResult.ModifiedCount, nil
}

func (r *MongoStore) DeleteByKey(ctx context.Context, key interface{}, itemType dal.Item) (int64, error) {
	collection := r.client.Database(itemType.Namespace()).Collection(itemType.ItemGroup())
	filter := bson.M{"_id": key}
	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("deleting entity: %w", err)
	}
	return deleteResult.DeletedCount, nil
}
