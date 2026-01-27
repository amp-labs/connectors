package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// nolint:lll
const (
	// See `Limiting Result Rows` section on this webpage:
	// https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select_fields.htm
	identifiersLimitStr = "200"
)

// SOQLBuilder builder of Salesforce Object Query Language.
// It constructs query dynamically.
type SOQLBuilder struct {
	fields string
	from   string
	where  []string
	limit  string
}

func (s *SOQLBuilder) SelectFields(fields []string) *SOQLBuilder {
	if slices.Contains(fields, "*") {
		s.fields = "FIELDS(ALL)"
		// if all fields are to be returned then we must limit to avoid error.
		// Error example: `The SOQL FIELDS function must have a LIMIT of at most 200`
		s.limit = identifiersLimitStr

		return s
	}

	s.fields = strings.Join(fields, ",")

	return s
}

func (s *SOQLBuilder) From(from string) *SOQLBuilder {
	s.from = from

	return s
}

func (s *SOQLBuilder) Limit(l int) *SOQLBuilder {
	s.limit = strconv.Itoa(l)

	return s
}

func (s *SOQLBuilder) Where(condition string) *SOQLBuilder {
	if s.where == nil {
		s.where = make([]string, 0)
	}

	s.where = append(s.where, condition)

	return s
}

func (s *SOQLBuilder) WithIDs(identifiers []string) *SOQLBuilder {
	// Decorate each id with quotes.
	for index, id := range identifiers {
		identifiers[index] = fmt.Sprintf("'%v'", id)
	}

	identifiersList := strings.Join(identifiers, ",")

	// nolint:lll
	// https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select_fields.htm
	return s.Where(fmt.Sprintf("Id IN (%v)", identifiersList))
}

func (s *SOQLBuilder) String() string {
	query := fmt.Sprintf("SELECT %s FROM %s", s.fields, s.from)

	if len(s.where) != 0 {
		query += " WHERE " + strings.Join(s.where, " AND ")
	}

	if len(s.limit) != 0 {
		query += " LIMIT " + s.limit
	}

	return query
}
