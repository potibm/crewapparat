package demozoo

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Scener is one row of the final, joined export.
type Scener struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	CountryCode string   `json:"country_code"`
	Groups      []string `json:"groups"`
}

type membershipRow struct {
	memberID  int64
	groupID   int64
	isCurrent bool
}

type dumpParser struct {
	groups       map[int64]string
	sceners      map[int64]*Scener
	memberships  []membershipRow
	inCopy       bool
	currentTable string
	colIndex     map[string]int
}

var copyHeaderRe = regexp.MustCompile(`(?i)^COPY\s+(\S+)\s*\((.*)\)\s+FROM\s+stdin;?\s*$`)

// ParseDump reads a PostgreSQL dump and returns all sceners (releasers with
// is_group = false), enriched with their current group memberships.
//
// The parser is intentionally tolerant of:
//   - reordered COPY sections,
//   - additional columns (even before the ones we use),
//   - reordered columns within a COPY header.
//
// Columns are resolved by name from each COPY header.
func ParseDump(r io.Reader) ([]Scener, error) {
	p := &dumpParser{
		groups:  map[int64]string{},
		sceners: map[int64]*Scener{},
	}

	reader := bufio.NewReader(r)

	for {
		line, err := reader.ReadString('\n')

		if err != nil && line == "" {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if line == "" {
			continue
		}

		trimmed := strings.TrimRight(line, "\r\n")
		if herr := p.handleLine(trimmed); herr != nil {
			return nil, herr
		}

		if err != nil {
			if err != io.EOF {
				return nil, err
			}

			break
		}
	}

	joinMemberships(p.sceners, p.groups, p.memberships)

	result := make([]Scener, 0, len(p.sceners))
	for _, s := range p.sceners {
		result = append(result, *s)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func (p *dumpParser) handleLine(trimmed string) error {
	switch {
	case !p.inCopy:
		table, idx, ok, err := parseCopyHeader(trimmed)
		if err != nil {
			return err
		}

		if ok && (table == "demoscene_releaser" || table == "demoscene_membership") {
			if verr := validateRequiredColumns(table, idx); verr != nil {
				return verr
			}

			p.currentTable = table
			p.colIndex = idx
			p.inCopy = true
		}
	case strings.TrimSpace(trimmed) == `\.`:
		p.inCopy = false
		p.currentTable = ""
		p.colIndex = nil
	case p.currentTable != "":
		fields := strings.Split(trimmed, "\t")
		processCopyDataRow(p.currentTable, fields, p.colIndex, p.groups, p.sceners, &p.memberships)
	}

	return nil
}

func parseCopyHeader(line string) (table string, colIndex map[string]int, ok bool, err error) {
	matches := copyHeaderRe.FindStringSubmatch(line)
	if matches == nil {
		return "", nil, false, nil
	}

	table = baseTableName(matches[1])
	if table != "demoscene_releaser" && table != "demoscene_membership" {
		return table, nil, false, nil
	}

	cols := splitColumns(matches[2])

	idx := make(map[string]int, len(cols))
	for i, c := range cols {
		idx[normalizeColumn(c)] = i
	}

	return table, idx, true, nil
}

func validateRequiredColumns(table string, idx map[string]int) error {
	var required []string

	switch table {
	case "demoscene_releaser":
		required = []string{"id", "name", "is_group"}
	case "demoscene_membership":
		required = []string{"member_id", "group_id", "is_current"}
	}

	for _, req := range required {
		if _, ok := idx[req]; !ok {
			return fmt.Errorf("%s COPY missing required column %q", table, req)
		}
	}

	return nil
}

func processCopyDataRow(
	currentTable string,
	fields []string,
	colIndex map[string]int,
	groups map[int64]string,
	sceners map[int64]*Scener,
	memberships *[]membershipRow,
) {
	if len(fields) != len(colIndex) {
		return
	}

	switch currentTable {
	case "demoscene_releaser":
		parseReleaserRow(fields, colIndex, groups, sceners)
	case "demoscene_membership":
		if m, ok := parseMembershipRow(fields, colIndex); ok {
			*memberships = append(*memberships, m)
		}
	}
}

func parseReleaserRow(fields []string, colIndex map[string]int, groups map[int64]string, sceners map[int64]*Scener) {
	id, err := strconv.ParseInt(unescape(fields[colIndex["id"]]), 10, 64)
	if err != nil {
		return
	}

	name := unescape(fields[colIndex["name"]])
	isGroup := unescape(fields[colIndex["is_group"]]) == "t"

	countryCode := ""
	if idx, ok := colIndex["country_code"]; ok {
		countryCode = unescape(fields[idx])
	}

	if isGroup {
		groups[id] = name

		return
	}

	sceners[id] = &Scener{
		ID:          id,
		Name:        name,
		CountryCode: countryCode,
		Groups:      []string{},
	}
}

func parseMembershipRow(fields []string, colIndex map[string]int) (membershipRow, bool) {
	memberID, err1 := strconv.ParseInt(unescape(fields[colIndex["member_id"]]), 10, 64)

	groupID, err2 := strconv.ParseInt(unescape(fields[colIndex["group_id"]]), 10, 64)
	if err1 != nil || err2 != nil {
		return membershipRow{}, false
	}

	isCurrent := unescape(fields[colIndex["is_current"]]) == "t"
	if !isCurrent {
		return membershipRow{}, false
	}

	return membershipRow{
		memberID:  memberID,
		groupID:   groupID,
		isCurrent: isCurrent,
	}, true
}

func joinMemberships(sceners map[int64]*Scener, groups map[int64]string, memberships []membershipRow) {
	for _, m := range memberships {
		scener, ok := sceners[m.memberID]
		if !ok {
			continue
		}

		groupName, ok := groups[m.groupID]
		if !ok {
			continue
		}

		if scener.Groups == nil {
			scener.Groups = []string{}
		}

		seen := make(map[string]struct{}, len(scener.Groups))
		for _, g := range scener.Groups {
			seen[g] = struct{}{}
		}

		if _, dup := seen[groupName]; !dup {
			scener.Groups = append(scener.Groups, groupName)
		}
	}
}

func baseTableName(name string) string {
	name = strings.TrimSpace(name)

	name = strings.Trim(name, `"`)
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}

	return strings.Trim(name, `"`)
}

func normalizeColumn(name string) string {
	return strings.ToLower(strings.Trim(strings.TrimSpace(name), `"`))
}

func splitColumns(s string) []string {
	parts := strings.Split(s, ",")

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}

	return out
}

const escapeAdvance = 2

// unescape converts a single field from PostgreSQL's COPY text format.
func unescape(s string) string {
	if s == `\N` {
		return ""
	}

	var b strings.Builder

	i := 0
	for i < len(s) {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			i++

			continue
		}

		repl, ni := unescapeOne(s, i)
		b.WriteString(repl)

		i = ni
	}

	return b.String()
}

func unescapeOne(s string, i int) (repl string, ni int) {
	if i+1 >= len(s) {
		return "\\", i + 1
	}

	switch s[i+1] {
	case '\\':
		return "\\", i + escapeAdvance
	case 't':
		return "\t", i + escapeAdvance
	case 'n':
		return "\n", i + escapeAdvance
	case 'r':
		return "\r", i + escapeAdvance
	case 'b':
		return "\b", i + escapeAdvance
	case 'f':
		return "\f", i + escapeAdvance
	case 'v':
		return "\v", i + escapeAdvance
	default:
		if r, n, ok := tryOctalEscape(s, i); ok {
			return r, n
		}

		return "\\", i + 1
	}
}

func tryOctalEscape(s string, i int) (repl string, ni int, ok bool) {
	if s[i+1] < '0' || s[i+1] > '7' {
		return "", 0, false
	}

	j := i + 1

	end := j
	for end < len(s) && end < j+3 && s[end] >= '0' && s[end] <= '7' {
		end++
	}

	if end == j {
		return "", 0, false
	}

	val, err := strconv.ParseUint(s[j:end], 8, 8)
	if err != nil {
		return "", 0, false
	}

	return string([]byte{byte(val)}), end, true
}
