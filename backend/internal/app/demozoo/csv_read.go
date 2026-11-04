package demozoo

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReadCSV reads a CSV produced by WriteCSV.
func ReadCSV(path string) ([]Scener, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil
	}

	// Skip header row.
	start := 0
	if len(rows[0]) > 0 && rows[0][0] == "id" {
		start = 1
	}

	result := make([]Scener, 0, len(rows)-start)
	for i := start; i < len(rows); i++ {
		row := rows[i]

		const demozooCSVColumns = 4

		if len(row) < demozooCSVColumns {
			continue
		}

		id, err := strconv.ParseInt(strings.TrimSpace(row[0]), 10, 64)
		if err != nil {
			continue
		}

		groups := []string{}
		if strings.TrimSpace(row[3]) != "" {
			groups = strings.Split(row[3], ";")
		}

		result = append(result, Scener{
			ID:          id,
			Name:        row[1],
			CountryCode: row[2],
			Groups:      groups,
		})
	}

	if len(result) == 0 && len(rows) > 1 {
		return nil, fmt.Errorf("no valid sceners found in %s", path)
	}

	return result, nil
}
