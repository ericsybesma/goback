package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/seebasoft/prompter/goback/dal"
	"github.com/seebasoft/prompter/goback/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

// Implement handlers that correspond to all of the Store options

// Provide a CRUD interface for dal.Item, enabling a REST API for
// any entity implementing this interface.
func Create(c *gin.Context, item dal.Item) {

	if err := c.ShouldBindJSON(item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dalStore := database.NewMongoStore(dbClient)
	item.SetKey(primitive.NewObjectID())
	objectID, err := dalStore.Create(c.Request.Context(), item)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.SetKey(objectID)
	c.JSON(http.StatusCreated, item)
}

func ReadByKey(c *gin.Context, item dal.Item) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	dalStore := database.NewMongoStore(dbClient)
	err = dalStore.ReadByKey(c.Request.Context(), objectID, item)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func ReadByFilter(c *gin.Context, item dal.Item) {
	ctx := c.Request.Context()
	queryOptions, err := ExtractQueryOptions(c, item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dalStore := database.NewMongoStore(dbClient)
	iter, err := dalStore.ReadByFilter(ctx, queryOptions, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

    defer iter.Close(ctx) // Close the cursor on exit
    entities := make([]dal.Item, 0) // Initialize entities slice
	for iter.Next(ctx) {
		ent := item.New()
		err := iter.Decode(ent)
		if err != nil {
			log.Fatal(err)
		}
		entities = append(entities, ent)
	}

	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
	c.JSON(http.StatusOK, entities)
}

func ExtractQueryOptions(c *gin.Context, item dal.Item) (queryOptions dal.QueryOptions, err error) {
	var filter bson.M
	queryOptions = database.NewMongoDalQueryOptions(database.NewMongoFilter(filter), bson.D{}, 0, 0)
	entityType, shouldReturn, err := getEntityType(item)
	if shouldReturn {
		return queryOptions, err
	}

	query := c.Request.URL.Query()
	if len(query) == 0 {
		return 	queryOptions, nil
	}
	sortOptions := getSortOptions(query)
	pageSize, page := getPagination(query)
	shouldReturn, filter, err = createFilter(entityType, query)
	if ! shouldReturn {
		queryOptions = database.NewMongoDalQueryOptions(database.NewMongoFilter(filter), sortOptions, pageSize, (page-1)*pageSize)
	}
	
	return queryOptions, err
}

// getSortOptions parses the "sort" query parameters and constructs a MongoDB sort option (bson.D)
//
// Example:
// Given a query with sort parameters: ?sort=-birthdate&sort=username
// The function will return a bson.D equivalent to: bson.D{{Key: "birthdate", Value: -1}, {Key: "username", Value: 1}}
func getSortOptions(query url.Values) bson.D {
    sortOptions := bson.D{} // Initialize to an empty bson.D
    sortFields := query["sort"]

    for _, field := range sortFields {
        order := 1
        if strings.HasPrefix(field, "-") {
            order = -1
            field = strings.TrimPrefix(field, "-")
        }
        sortOptions = append(sortOptions, bson.E{Key: field, Value: order})
    }

    return sortOptions
}

func getPagination(query url.Values) (pageSize int64, page int64) {
	pageSize = 10 // Default page size
	page = 1 // Default page number

	if size := query.Get("pageSize"); size != "" {
		pageSizeInt, err := strconv.Atoi(size)
		if err != nil {
			return pageSize, page
		}
		pageSize = int64(pageSizeInt)
	}

	if p := query.Get("page"); p != "" {
		pageInt, err := strconv.Atoi(p)
		if err == nil {
			page = int64(pageInt)
		}
	}
	return pageSize, page
}

func getEntityType(item dal.Item) (entityType reflect.Type, shouldReturn bool, err error) {
	entityValue := reflect.ValueOf(item)
	if entityValue.Kind() == reflect.Ptr {
		entityValue = entityValue.Elem()
	}

	entityType = entityValue.Type()
	if entityType.Kind() != reflect.Struct {
		err = fmt.Errorf("item must be a struct or pointer to struct")
		shouldReturn = true
	}
	return entityType, shouldReturn, err
}

func createFilter(entityType reflect.Type, query url.Values) (shouldReturn bool, filter bson.M, err error) {
	filter = bson.M{}
	for fieldName := 0; fieldName < entityType.NumField(); fieldName++ {
		field := entityType.Field(fieldName)
		shouldSkip, jsonName, bsonName := getTagNames(field)
		if shouldSkip {
			continue
		}

		for param, values := range query {
			if len(values) == 0 || strings.Split(param, "_")[0] != jsonName{
				continue
			}

			subFilter, shouldReturn, err := paramToSubFilter(field, param, jsonName, values[0])
			if shouldReturn {
				return true, filter, err
			}

			if existingFilter, ok := filter[bsonName]; ok {
				if andFilter, ok := existingFilter.(bson.M)["$and"]; ok {
					filter[bsonName] = bson.M{"$and": append(andFilter.([]bson.M), subFilter)}
				} else {
					filter[bsonName] = bson.M{"$and": []bson.M{existingFilter.(bson.M), subFilter}}
				}
			} else {
				filter[bsonName] = subFilter
			}
		}
	}

	return false, filter, err
}

func getTagNames(field reflect.StructField) (shouldSkip bool, jsonName string, bsonName string) {
	jsonTag := field.Tag.Get("json")
	bsonTag := field.Tag.Get("bson")
	if jsonTag == "" || bsonTag == "" {
		return true, "", ""
	}
	jsonName = strings.Split(jsonTag, ",")[0]
	bsonName = strings.Split(bsonTag, ",")[0]
	return false, jsonName, bsonName
}

func paramToSubFilter(field reflect.StructField, param string, jsonName string, strval string) (subFilter bson.M, shouldReturn bool, err error) {
	// adjust a param without an underscore to have an "_eq" suffix
	if param == jsonName {
		param = jsonName + "_eq"
	}

	var value interface{} = strval
	if field.Type == reflect.TypeOf(time.Time{}) {
		t, err := time.Parse(time.RFC3339, strval)
		if err != nil {
			return nil, true, fmt.Errorf("invalid date format for %s: %w", param, err)
		}
		value = t
	} else if field.Type.Kind() == reflect.Float32 || field.Type.Kind() == reflect.Float64 {
		f, err := strconv.ParseFloat(strval, 64)
		if err != nil {
			return nil, true, fmt.Errorf("invalid float format for %s: %w", param, err)
		}
		value = f
	}
	
	op := strings.TrimPrefix(param, jsonName+"_")
	subFilter = bson.M{}

	switch op {
	case "eq":
		subFilter["$eq"] = value
	case "ne":
		subFilter["$ne"] = value
	case "gt", "after":
		subFilter["$gt"] = value
	case "gte":
		subFilter["$gte"] = value
	case "lt", "before":
		subFilter["$lt"] = value
	case "lte":
		subFilter["$lte"] = value
	case "contains":
		subFilter["$regex"] = primitive.Regex{Pattern: regexp.QuoteMeta(strval), Options: "i"}
	case "startswith":
		subFilter["$regex"] = primitive.Regex{Pattern: "^" + regexp.QuoteMeta(strval), Options: "i"}
	case "endswith":
		subFilter["$regex"] = primitive.Regex{Pattern: regexp.QuoteMeta(strval) + "$", Options: "i"}
	case "between":
		parts := strings.Split(strval, ",")
		if len(parts) != 2 {
			return nil, true, fmt.Errorf("invalid between value for %s: %s", param, value)
		}
		start, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, true, fmt.Errorf("invalid start date format for %s: %w", param, err)
		}
		end, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, true, fmt.Errorf("invalid end date format for %s: %w", param, err)
		}
		subFilter["$gte"] = start
		subFilter["$lte"] = end
	default:
		return nil, true, fmt.Errorf("invalid filter operator: %s", op)
	}
	return subFilter, false, nil
}

func UpdateByKey(c *gin.Context, item dal.Item) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := c.ShouldBindJSON(item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.SetKey(objectID)
	dalStore := database.NewMongoStore(dbClient)
	_, err = dalStore.UpdateByKey(c.Request.Context(), objectID, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func DeleteByKey(c *gin.Context, item dal.Item) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Save the result to get the DeletedCount
	dalStore := database.NewMongoStore(dbClient)
	deletedCount, err := dalStore.DeleteByKey(c.Request.Context(), objectID, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if deletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No entity found to delete"})
		return
	}

	label := "entries"
	if deletedCount == 1 {
		label = "entry"
	}
	msg := fmt.Sprintf("Deleted %d %s", deletedCount, label)
	c.JSON(http.StatusOK, gin.H{"message": msg})
}
