package salesforce

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// nolint:lll
const (
	// See `Limiting Result Rows` section on this webpage:
	// https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select_fields.htm
	identifiersLimit    = 200
	identifiersLimitStr = "200"
)

// soqlBuilder builder of Salesforce Object Query Language.
// It constructs query dynamically.
type soqlBuilder struct {
	fields string
	from   string
	where  []string
	limit  string
}

func (s *soqlBuilder) SelectFields(fields []string) *soqlBuilder {
	for _, field := range fields {
		if field == "*" {
			s.fields = "FIELDS(ALL)"
			// if all fields are to be returned then we must limit to avoid error.
			// Error example: `The SOQL FIELDS function must have a LIMIT of at most 200`
			s.limit = identifiersLimitStr

			return s
		}
	}

	s.fields = strings.Join(fields, ",")

	return s
}

func (s *soqlBuilder) From(from string) *soqlBuilder {
	s.from = from

	return s
}

func (s *soqlBuilder) Where(condition string) *soqlBuilder {
	if s.where == nil {
		s.where = make([]string, 0)
	}

	s.where = append(s.where, condition)

	return s
}

func (s *soqlBuilder) WithIDs(identifiers []string) error {
	if len(identifiers) > identifiersLimit {
		return common.ErrTooManyRecordIDs
	}

	// Decorate each id with quotes.
	for index, id := range identifiers {
		identifiers[index] = fmt.Sprintf("'%v'", id)
	}

	identifiersList := strings.Join(identifiers, ",")

	// nolint:lll
	// https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select_fields.htm
	s.Where(fmt.Sprintf("Id IN (%v)", identifiersList))

	return nil
}

func (s *soqlBuilder) String() string {
	query := fmt.Sprintf("SELECT %s FROM %s", s.fields, s.from)

	if len(s.where) != 0 {
		query += " WHERE " + strings.Join(s.where, " AND ")
	}

	if len(s.limit) != 0 {
		query += " LIMIT " + s.limit
	}

	return query
}
