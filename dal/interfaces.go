package dal

import (
    "fmt"
	"context"
)

type Key interface {
    Value() interface{} // Returns the key value
    String() string      // Returns string representation
}

// AnyKey is a default implementation of Key that wraps any value.
type AnyKey struct {
    value interface{}
}

// NewAnyKey creates a new AnyKey.
func NewAnyKey(v interface{}) Key {
    return AnyKey{value: v}
}

func (k AnyKey) Value() interface{} { return k.value }
func (k AnyKey) String() string {
    if s, ok := k.value.(string); ok {
            return s
    }
    return fmt.Sprintf("%v", k.value) // Default string conversion
}

// Filter represents a database query filter.
type Filter interface {
	ToNative() interface{} // Converts the filter to a database-specific format
}

// QueryOptions for filtering, sorting, pagination.
type QueryOptions interface {
	GetFilter() Filter
	GetSort() interface{} // Database-specific sort representation
	GetLimit() int64
	GetSkip() int64
}

// Item represents a data entity.
type Item interface {
	Namespace() string
	ItemGroup() string
	GetKey() interface{}
	SetKey(interface{}) error
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	New() Item
}

// ItemIterator iterates over query results.
type ItemIterator interface {
	Next(ctx context.Context) bool
	Decode(item Item) error
	Close(ctx context.Context) error
	Err() error
}

// Store provides data access operations.
type Store interface {
	Create(ctx context.Context, item Item) (Item, error)
	ReadByKey(ctx context.Context, key interface{}, item Item) error
	ReadByFilter(ctx context.Context, options QueryOptions, itemType Item) (ItemIterator, error)
	UpdateByKey(ctx context.Context, key interface{}, item Item) (int64, error)
	DeleteByKey(ctx context.Context, key interface{}, itemType Item) (int64, error)
}
