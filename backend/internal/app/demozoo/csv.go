package demozoo

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
)

// WriteCSV writes the sceners as CSV with a header row. Group names inside the
// "groups" field are separated by semicolons.
func WriteCSV(w io.Writer, sceners []Scener) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "name", "country_code", "groups"}); err != nil {
		return err
	}

	for _, s := range sceners {
		if err := cw.Write([]string{
			strconv.FormatInt(s.ID, 10),
			s.Name,
			s.CountryCode,
			strings.Join(s.Groups, ";"),
		}); err != nil {
			return err
		}
	}

	cw.Flush()

	return cw.Error()
}
