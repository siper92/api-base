package api_entity

import "gorm.io/gorm"

type ToModelEntity[T any] interface {
	ToModel() T
}

type RepoEntity interface {
	GetID() int64
	TableName() string
	IsFilterable(field string) bool
}

type CollectionFilter[T any] interface {
	Condition() string
	Values() []any
	ApplyTo(c T) (T, error)
}

type GormFilter interface {
	CollectionFilter[*gorm.DB]
}

type EntityRepository[T RepoEntity] interface {
	New() T
	NewSlice() []T
	GetByID(id int64) (T, error)
	Create(T) (T, error)
	Update(T) (bool, error)
	Delete(T) (bool, error)
	Select() *gorm.DB
	ApplyFilters(filters ...GormFilter) ([]T, error)
}
