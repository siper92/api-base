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
	ApplyFilters(filters ...GormFilter) *gorm.DB
	GetOne(filters ...GormFilter) (T, error)
	GetResults(filters ...GormFilter) ([]T, error)
	Count(filters ...GormFilter) (int64, error)

	Create(T) (T, error)
	Update(T) (bool, error)
	Delete(T) (bool, error)
}
