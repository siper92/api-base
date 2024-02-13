package api_entity

import (
	"fmt"
	"github.com/siper92/api-base"
	"github.com/siper92/core-utils/type_utils"
	"gorm.io/gorm"
)

type GormFilter interface {
	api_base.CollectionFilter[*gorm.DB]
}

var _ GormFilter = (WhereFilter)(nil)

type WhereFilter []any

func (w WhereFilter) Values() []any {
	if len(w) > 1 {
		return w[1:]
	}

	return nil
}

func toStringVal(w any) string {
	switch val := w.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		panic("invalid where filter value type: " + fmt.Sprintf("%T", val) + " " + fmt.Sprintf("%v", w))
	}
}

func (w WhereFilter) Condition() string {
	if len(w) == 0 {
		panic("empty where filter")
	}

	return toStringVal(w[0])
}

func (w WhereFilter) ApplyTo(c *gorm.DB) (*gorm.DB, error) {
	values := w.Values()
	if len(values) < 1 {
		return c.Where(w.Condition()), nil
	}

	return c.Where(w.Condition(), values...), nil
}

var _ GormFilter = (*PageFilter)(nil)

type PageFilter struct {
	Page  int
	Limit int
}

func (p PageFilter) Condition() string {
	if p.Limit == 0 {
		return ""
	}

	if p.Page == 0 {
		return fmt.Sprintf("LIMIT %d", p.Limit)
	}

	return fmt.Sprintf("LIMIT %d OFFSET %d", p.Limit, p.Page*p.Limit)
}

func (p PageFilter) Values() []any {
	return []any{p.Limit, p.Page * p.Limit}
}

func (p PageFilter) ApplyTo(c *gorm.DB) (*gorm.DB, error) {
	return c.Limit(p.Limit).Offset(p.Page * p.Limit), nil
}

type FilterType string

const (
	EQ FilterType = "eq"
	NE FilterType = "ne"
	GE FilterType = "ge"
	GT FilterType = "gt"
	LE FilterType = "le"
	LT FilterType = "lt"
)

var AllFilterTypes = []FilterType{EQ, GE, GT, LE, LT, NE}

var _ GormFilter = (*FilterField)(nil)

type FilterField struct {
	Field string
	Type  FilterType
	Value any
}

func (f FilterField) Condition() string {
	for _, val := range AllFilterTypes {
		if val == f.Type {
			return fmt.Sprintf("%s %s ?", f.Field, f.Type)
		}
	}

	panic(fmt.Sprintf("unsuported condition type: %s", f.Type))
}

func (f FilterField) Values() []any {
	val := type_utils.BaseTypeToString(f.Value)
	if val == "" {
		return nil
	}

	return []any{val}
}

func (f FilterField) ApplyTo(c *gorm.DB) (*gorm.DB, error) {
	return WhereFilter{f.Condition(), f.Values()}.ApplyTo(c)
}
