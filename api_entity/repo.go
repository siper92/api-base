package api_entity

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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
	ApplyTo(c *gorm.DB) (*gorm.DB, error)
}

type GormRepository[T RepositoryItem] interface {
	Conn() *gorm.DB
	New() T
	NewSlice() []T

	GetByID(id int64) (T, error)
	GetByIDs(ids ...int64) ([]T, error)
	GetResults(filters ...GormFilter) ([]T, error)
	ApplyFilters(filters ...GormFilter) *gorm.DB
	Count(filters ...GormFilter) (int64, error)

	Create(T) (T, error)
	Update(T) (bool, error)
	Delete(T) (bool, error)
}
