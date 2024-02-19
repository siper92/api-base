package repo_filter

import (
	"github.com/siper92/api-base/api_entity"
	"testing"
)

func Test_PrepareFilters(t *testing.T) {
	t.Run("Simple case", func(t *testing.T) {
		filters := PrepareFilters("active = ?", 1)
		if len(filters) != 1 {
			t.Errorf("Expected 1 filter, got %d", len(filters))
		} else if filters[0].Condition() != "active = ?" {
			t.Errorf("Expected 'active = ?', got %s", filters[0].Condition())
		} else if filters[0].Values()[0] != 1 {
			t.Errorf("Expected 1, got %d", filters[0].Values()[0])
		}
	})

	t.Run("Panic on wrong number of args", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("PrepFilters() should have panicked")
			}
		}()

		_ = PrepareFilters("active = 2", 1)
	})

	testsErrors := []struct {
		name          string
		filter        string
		gormFilter    api_entity.GormFilter
		values        []interface{}
		expectedCount int
		errorMessage  string
	}{
		{
			name:          "Simple case",
			filter:        "active = ? or value = ?",
			values:        []interface{}{1},
			expectedCount: 2,
		},
		{
			name:          "Multiple values",
			filter:        "active = ? or value = ? or id = ?",
			values:        []interface{}{1},
			expectedCount: 3,
		},
		{
			name:          "GormFilter",
			gormFilter:    Raw("active = ? or value = ?"),
			values:        []interface{}{1, 2, 3, 4},
			expectedCount: 2,
		},
	}

	for _, tt := range testsErrors {
		t.Run(tt.name, func(t *testing.T) {
			filterCon := tt.filter
			if filterCon == "" && tt.gormFilter != nil {
				filterCon = tt.gormFilter.Condition()
			}

			defer func() {
				err := recover().(string)
				expected := getErrorMessage(filterCon, tt.expectedCount, len(tt.values))
				if tt.errorMessage != "" {
					expected = tt.errorMessage
				}

				if err == "" {
					t.Errorf("PrepareFilters() should have panicked")
				} else if err != expected {
					t.Errorf("mistmatched errors \n - %s, \n - %s", expected, err)
				}
			}()

			var filters []interface{}
			if tt.gormFilter != nil {
				filters = []interface{}{
					filterCon,
				}
			} else {
				filters = []interface{}{
					filterCon,
				}
			}

			filters = append(filters, tt.values...)
			_ = PrepareFilters(filters...)
		})
	}
}
