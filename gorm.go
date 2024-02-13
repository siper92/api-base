package api_base

import (
	"fmt"
	"gorm.io/gorm"
)

type GormFilter interface {
	CollectionFilter[*gorm.DB]
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
	return c.Where(w.Condition(), w.Values()...), nil
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
