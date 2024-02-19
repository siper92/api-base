package api_entity

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type NotFilterableError struct {
	Field string
}

func (e NotFilterableError) Error() string {
	return "field is not filterable: " + e.Field
}

type ConvertToModel[T any] interface {
	ToModel() T
}

type RepositoryItem interface {
	schema.Tabler
	GetID() int64
}

type RepositoryEntity interface {
	IsFilterable(field string) bool
}

type GormFilter interface {
	Condition() string
	Values() []interface{}
	ApplyTo(c *gorm.DB) *gorm.DB
}

type GormRepository[T RepositoryItem] interface {
	Conn() *gorm.DB
	New() T
	NewSlice() []T

	GetByID(id int64) (T, error)
	GetByIDs(ids ...int64) ([]T, error)
	ApplyFilters(filters ...interface{}) *gorm.DB
	GetOne(filters ...interface{}) (T, error)
	GetResults(filters ...interface{}) ([]T, error)
	Count(filters ...interface{}) (int64, error)

	Create(T) (T, error)
	Update(T) (bool, error)
	Delete(T) (bool, error)
}
