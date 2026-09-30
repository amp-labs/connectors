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
//
// Check the "SOQL and SOSL Reference" book:
// https://resources.docs.salesforce.com/latest/latest/en-us/sfdc/pdf/salesforce_soql_sosl.pdf
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

func (s *SOQLBuilder) Limit(limit int64) *SOQLBuilder {
	s.limit = strconv.FormatInt(limit, 10)

	return s
}

func (s *SOQLBuilder) Where(condition string) *SOQLBuilder {
	if s.where == nil {
		s.where = make([]string, 0)
	}

	s.where = append(s.where, condition)

	return s
}

// WhereRawFilter adds a raw filter condition wrapped in parentheses, so the condition's own
// OR/AND grouping cannot bind with the other conditions, which are joined by AND.
func (s *SOQLBuilder) WhereRawFilter(condition string) *SOQLBuilder {
	return s.Where("(" + condition + ")")
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

// SelectCount sets the query to use COUNT() instead of field selection.
func (s *SOQLBuilder) SelectCount() *SOQLBuilder {
	s.fields = "COUNT()"

	return s
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

// QueryEndpoint returns the REST resource that runs a SOQL query: queryAll when the query
// must be able to return deleted and archived records, query otherwise.
// The query resource never returns deleted or archived records, whatever the WHERE clause says,
// so reads and searches with a raw filter use queryAll and let the filter decide which of those
// are returned.
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_queryall.htm
func QueryEndpoint(includeDeletedAndArchived bool) string {
	if includeDeletedAndArchived {
		return "queryAll"
	}

	return "query"
}
