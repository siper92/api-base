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

type Entity interface {
	schema.Tabler
	GetID() int64
}

type Initializable[T Entity] interface {
	New() T
	NewSlice() []T
}

type GormFilter interface {
	Condition() string
	Values() []interface{}
	ApplyTo(c *gorm.DB) *gorm.DB
}

type HasRepository[T Entity] interface {
	NewRepository() Repository[T]
}

type Repository[T Entity] interface {
	Conn() *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB // table name added to select

	Filter(filters ...GormFilter) *gorm.DB
	GetByID(id int64) (T, error)
	GetByIDs(ids ...int64) ([]T, error)
	GetOne(filters ...GormFilter) (T, error)
	GetResults(filters ...GormFilter) ([]T, error)
}

type Creator[T Entity] interface {
	Create(T) (T, error)
}

type Updater[T Entity] interface {
	Update(T) (bool, error)
}

type Remover[T Entity] interface {
	Delete(T) (bool, error)
}
