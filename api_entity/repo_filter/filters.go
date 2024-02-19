package repo_filter

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
)

func toCondition(w any) string {
	switch val := w.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		panic("invalid where condition type: " + fmt.Sprintf("%T", w))
	}
}

type FilterType string

type Where []any

func (w Where) Condition() string {
	if len(w) == 0 {
		panic("Empty where")
	}

	return toCondition(w[0])
}

func (w Where) Values() []any {
	if len(w) < 2 {
		return nil
	}

	values := w[1:]
	valuesCount := strings.Count(w.Condition(), "?")
	if len(values) != valuesCount {
		panic("invalid number of parameters for: " + w.Condition())
	}

	return w[1:]
}

func (w Where) ApplyTo(c *gorm.DB) *gorm.DB {
	if len(w) > 1 {
		return c.Where(w.Condition(), w.Values()...)
	}

	return c.Where(w.Condition())
}

type Raw string

func (r Raw) Condition() string {
	return string(r)
}

func (r Raw) Values() []interface{} {
	return nil
}

func (r Raw) ApplyTo(c *gorm.DB) *gorm.DB {
	return Where{r.Condition()}.ApplyTo(c)
}
