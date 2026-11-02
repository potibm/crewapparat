package gorm

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/potibm/crewapparat/internal/app/repository"
)

// applySort allows only the columns listed in allowed. The value comes from the
// query string (`?_sort=`) and would otherwise reach ORDER BY unchecked, where
// SQLite accepts subqueries - turning a sort parameter into a data leak.
func applySort(
	query *gorm.DB,
	allowed map[string]string,
	requested string,
	order string,
) (*gorm.DB, error) {
	if requested == "" {
		requested = defaultSortColumn
	}

	column, ok := allowed[requested]
	if !ok {
		return nil, fmt.Errorf("unsupported sort column: %q", requested)
	}

	desc := order == "DESC"

	return query.Order(clause.OrderByColumn{
		Column: clause.Column{Name: column},
		Desc:   desc,
	}), nil
}

const defaultSortColumn = "id"

// applyPagination sets offset and limit. A limit of 0 means "no limit", so small
// reference tables can be loaded in full.
func applyPagination(query *gorm.DB, params repository.ListParams) *gorm.DB {
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	return query
}
