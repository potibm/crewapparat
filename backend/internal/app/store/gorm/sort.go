package gorm

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/potibm/crewapparat/internal/app/repository"
)

// sortColumns erlaubt ausschließlich die hier gelisteten Spalten als Sortierung.
// Der Wert stammt aus der Query (`?_sort=`) und landet sonst ungeprüft im
// ORDER BY - dort akzeptiert SQLite Unterabfragen, was aus einem
// Sortierparameter ein Datenleck machen würde.
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

// applyPagination setzt Offset und Limit. Limit 0 bedeutet "kein Limit",
// damit zum Beispiel kleine Referenztabellen komplett geladen werden können.
func applyPagination(query *gorm.DB, params repository.ListParams) *gorm.DB {
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	return query
}
