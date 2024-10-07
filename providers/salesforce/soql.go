package salesforce

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
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
			s.limit = "200"

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

// makeSOQL returns the SOQL query for the desired read operation.
func makeSOQL(config common.ReadParams) *soqlBuilder {
	soql := (&soqlBuilder{}).SelectFields(config.Fields.List()).From(config.ObjectName)

	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	if !config.Since.IsZero() {
		soql.Where("SystemModstamp > " + handy.Time.FormatRFC3339inUTC(config.Since))
	}

	if config.Deleted {
		soql.Where("IsDeleted = true")
	}

	// TODO: When we support builder facing filters, we should escape the
	// filter string to avoid SOQL injection.
	if config.Filter != "" {
		soql.Where(config.Filter)
	}

	return soql
}
