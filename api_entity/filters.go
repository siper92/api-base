package api_entity

import (
	"fmt"
	"github.com/siper92/core-utils/type_utils"
	"gorm.io/gorm"
)

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

func (f FilterType) ToCondition() string {
	return string(f)
}

const (
	EQ      FilterType = "="
	NE      FilterType = "!="
	GE      FilterType = ">="
	GT      FilterType = ">"
	LE      FilterType = "<="
	LT      FilterType = "<"
	LIKE    FilterType = "LIKE"
	IN      FilterType = "IN"
	NOTIN   FilterType = "NOT IN"
	BETWEEN FilterType = "BETWEEN"
)

var _ GormFilter = (*FilterField)(nil)

type FilterField struct {
	Field string
	Type  FilterType
	Value any
}

func (f FilterField) Condition() string {
	if f.Value == nil {
		return ""
	}

	if f.Type == IN || f.Type == NOTIN {
		return fmt.Sprintf("%s %s (?)", f.Field, f.Type)
	} else if f.Type == LIKE {
		return fmt.Sprintf("%s %s ?", f.Field, f.Type)
	}

	return fmt.Sprintf("%s %s ? AND ?", f.Field, f.Type)
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
