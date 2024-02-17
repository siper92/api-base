package api_entity

import (
	"fmt"
	core_utils "github.com/siper92/core-utils"
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

func (w WhereFilter) ApplyTo(c *gorm.DB) *gorm.DB {
	values := w.Values()
	if len(values) < 1 {
		return c.Where(w.Condition())
	}

	return c.Where(w.Condition(), values...)
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

func (p PageFilter) ApplyTo(c *gorm.DB) *gorm.DB {
	return c.Limit(p.Limit).Offset(p.Page * p.Limit)
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

var allFilterTypes = []FilterType{EQ, NE, GE, GT, LE, LT, LIKE, IN, NOTIN, BETWEEN}

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

	for _, ft := range allFilterTypes {
		if ft == f.Type {
			if f.Type == IN || f.Type == NOTIN {
				return fmt.Sprintf("%s %s (?)", f.Field, f.Type)
			} else if f.Type == BETWEEN {
				return fmt.Sprintf("%s %s ? AND ?", f.Field, f.Type)
			}

			return fmt.Sprintf("%s %s ?", f.Field, f.Type)
		}
	} // end for

	core_utils.Debug("invalid filter type: " + string(f.Type))
	return fmt.Sprintf("%s %s ?", f.Field, EQ)
}

func (f FilterField) Values() []any {
	val := type_utils.BaseTypeToString(f.Value)
	if val == "" {
		return nil
	}

	return []any{val}
}

func (f FilterField) ApplyTo(c *gorm.DB) *gorm.DB {
	return WhereFilter{f.Condition(), f.Values()}.ApplyTo(c)
}

func ApplyFilters(c *gorm.DB, filters ...any) (*gorm.DB, error) {
	var condition string
	var values []any

	for _, f := range filters {
		switch fType := f.(type) {
		case GormFilter:
			c = fType.ApplyTo(c)
		case string:
			if condition == "" {
				condition = fType
			} else {
				values = append(values, fType)
			}
		default:
			if condition == "" {
				return c, fmt.Errorf("invalid filter type: " + fmt.Sprintf("%T", f))
			} else {
				values = append(values, f)
			}
		}
	}

	if condition != "" {
		return c.Where(condition, values...), nil
	}

	return nil, fmt.Errorf("no filters provided or unknow types")
}
