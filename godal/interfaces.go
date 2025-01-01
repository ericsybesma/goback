package dal

import (
    "context"
)

// Key represents a database key.
type Key interface {
    Value() interface{} // Returns the key value
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

// dal.Item represents a data entity.
type Item interface {
    Namespace() string
    ItemGroup() string
    GetKey() Key
    SetKey(Key) error
    Marshal() ([]byte, error)
    Unmarshal([]byte) error
    New() Item
}

// dal.ItemIterator iterates over query results.
type ItemIterator interface {
    Next(ctx context.Context) bool
    Decode(item Item) error
    Close(ctx context.Context) error
    Err() error
}

// Store provides data access operations.
type Store interface {
    Create(ctx context.Context, item Item) (Item, error)
    ReadByKey(ctx context.Context, key Key, item Item) error
    ReadByFilter(ctx context.Context, options QueryOptions, itemType Item) (ItemIterator, error)
    UpdateByKey(ctx context.Context, key Key, item Item) (int64, error)
    DeleteByKey(ctx context.Context, key Key, itemType Item) (int64, error)
}
