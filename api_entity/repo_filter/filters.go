package repo_filter

import (
	"fmt"
	"github.com/siper92/api-base/api_entity"
	core_utils "github.com/siper92/core-utils"
	"github.com/siper92/core-utils/type_utils"
	"gorm.io/gorm"
	"strings"
)

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

var _ api_entity.GormFilter = (*Pager)(nil)

type Pager struct {
	Page  int
	Limit int
}

func (p Pager) Condition() string {
	if p.Limit == 0 {
		return ""
	}

	if p.Page == 0 {
		return fmt.Sprintf("LIMIT %d", p.Limit)
	}

	return fmt.Sprintf("LIMIT %d OFFSET %d", p.Limit, p.Page*p.Limit)
}

func (p Pager) Values() []any {
	return []any{p.Limit, p.Page * p.Limit}
}

func (p Pager) ApplyTo(c *gorm.DB) *gorm.DB {
	return c.Limit(p.Limit).Offset(p.Page * p.Limit)
}

type FilterType string

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
	NOTNULL FilterType = "IS NOT NULL"
)

var allFilterTypes = map[FilterType]bool{
	EQ:      true,
	NE:      true,
	GE:      true,
	GT:      true,
	LE:      true,
	LT:      true,
	LIKE:    true,
	IN:      true,
	NOTIN:   true,
	BETWEEN: true,
	NOTNULL: true,
}

var _ api_entity.GormFilter = (*Field)(nil)

type Field struct {
	Field string
	Type  FilterType
	Value any
}

func (f Field) Condition() string {
	if f.Value == nil {
		return ""
	}

	if _, ok := allFilterTypes[f.Type]; ok {
		if f.Type == IN || f.Type == NOTIN {
			return fmt.Sprintf("%s %s (?)", f.Field, f.Type)
		} else if f.Type == BETWEEN {
			return fmt.Sprintf("%s %s ? AND ?", f.Field, f.Type)
		}

		return fmt.Sprintf("%s %s ?", f.Field, f.Type)
	}

	core_utils.Debug("invalid filter type: " + string(f.Type))
	return fmt.Sprintf("%s %s ?", f.Field, EQ)
}

func (f Field) Values() []any {
	val := type_utils.BaseTypeToString(f.Value)
	if val == "" {
		return nil
	}

	return []any{val}
}

func (f Field) ApplyTo(c *gorm.DB) *gorm.DB {
	return Where{Cmd: f.Condition(), Value: f.Values()}.ApplyTo(c)
}

type Where struct {
	Cmd   string
	Value []any
}

func (w Where) Condition() string {
	return w.Cmd
}

func (w Where) Values() []interface{} {
	return w.Value
}

func (w Where) ApplyTo(c *gorm.DB) *gorm.DB {
	return c.Where(w.Condition(), w.Values()...)
}

type Raw string

func (r Raw) Condition() string {
	return string(r)
}

func (r Raw) Values() []interface{} {
	return nil
}

func (r Raw) ApplyTo(c *gorm.DB) *gorm.DB {
	return Where{Cmd: r.Condition(), Value: r.Values()}.ApplyTo(c)
}

func PrepareFilters(filters ...interface{}) []api_entity.GormFilter {
	preparedFilters := make([]api_entity.GormFilter, 0)
	var lastCondition api_entity.GormFilter

	for _, filter := range filters {
		switch filterPart := filter.(type) {
		case api_entity.GormFilter:
			if lastCondition != nil {
				preparedFilters = append(preparedFilters, lastCondition)
			}
			lastCondition = filterPart
		case string:
			if lastCondition != nil {
				preparedFilters = append(preparedFilters, lastCondition)
			}

			lastCondition = Where{
				Cmd: filterPart,
			}
		default:
			switch t := lastCondition.(type) {
			case Field:
				values := t.Values()
				valuesCount := strings.Count(t.Condition(), "?")
				if canAddValueToFilter(t) == false {
					panic(fmt.Sprintf("InvalidArgsNum: '%s': %d != %d", lastCondition.Condition(), valuesCount, len(values)))
				}

				lastCondition = Field{
					Field: t.Field,
					Type:  t.Type,
					Value: append(values, filterPart),
				}
			case Where:
				values := t.Values()
				valuesCount := strings.Count(t.Condition(), "?")
				if canAddValueToFilter(t) == false {
					getErrorMessage(t.Condition(), valuesCount, len(values))
				}

				lastCondition = Where{
					Cmd:   t.Cmd,
					Value: append(values, filterPart),
				}
			default:
				panic(fmt.Sprintf("unsupported filter type %T", filter))
			}
		}
	}

	if lastCondition != nil {
		valuesCount := strings.Count(lastCondition.Condition(), "?")
		actualValuesCount := len(lastCondition.Values())
		if valuesCount != actualValuesCount {
			panic(getErrorMessage(lastCondition.Condition(), valuesCount, actualValuesCount))
		}

		preparedFilters = append(preparedFilters, lastCondition)
	}

	return preparedFilters
}

func canAddValueToFilter(filter api_entity.GormFilter) bool {
	values := filter.Values()
	valuesCount := strings.Count(filter.Condition(), "?")
	return valuesCount >= len(values)+1
}

func getErrorMessage(cond string, expectedValCount int, actualValCount int) string {
	return fmt.Sprintf("InvalidArgsNum: '%s': expected %d got %d", cond, expectedValCount, actualValCount)
}
