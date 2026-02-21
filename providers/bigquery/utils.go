package bigquery

import (
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// readSessionToken holds the state for paginated reads using the Storage Read API.
// This token is serialized to JSON, base64-encoded, and returned as NextPage.
// It contains all information needed to resume reading from where we left off.
//
// # Multi-stream state tracking
//
// BigQuery Storage API creates multiple parallel streams for a read session.
// Each stream has its own offset and completion state. The token tracks all
// streams so that pagination can resume correctly across multiple Read calls.
type readSessionToken struct {
	// SessionName is the BigQuery Storage session identifier.
	// Sessions are server-managed and have a limited lifetime (~24 hours).
	SessionName string `json:"session"`

	// Streams contains the state of each parallel stream.
	// BigQuery splits the table into multiple streams for parallel reading.
	// Each stream tracks its own offset and done state independently.
	Streams []streamState `json:"streams"`

	// ActiveStreams is the count of streams that still have data to read.
	// When this reaches 0, we've exhausted all data and Done=true is returned.
	ActiveStreams int `json:"activeStreams"`
}

// streamState tracks the read position within a single BigQuery Storage stream.
type streamState struct {
	// Name is the fully-qualified stream resource name.
	Name string `json:"name"`

	// Offset is the number of rows already read from this stream.
	// Used to resume reading from the correct position.
	Offset int64 `json:"offset"`

	// Done indicates this stream has been fully consumed (EOF reached).
	Done bool `json:"done"`
}

// validateFieldCount enforces the MaxFields limit.
// Returns a descriptive error if the limit is exceeded.
func validateFieldCount(fields datautils.StringSet) error {
	fieldList := fields.List()

	if len(fieldList) > MaxFields {
		return fmt.Errorf("too many fields requested: %d (max %d). Reduce field count for better performance", len(fieldList), MaxFields)
	}

	return nil
}

// bigqueryTypeToValueType maps BigQuery field types to Ampersand ValueTypes.
func bigqueryTypeToValueType(bqType bigquery.FieldType) common.ValueType {
	switch bqType {
	case bigquery.StringFieldType:
		return common.ValueTypeString
	case bigquery.BytesFieldType:
		return common.ValueTypeString
	case bigquery.IntegerFieldType:
		return common.ValueTypeInt
	case bigquery.FloatFieldType:
		return common.ValueTypeFloat
	case bigquery.NumericFieldType, bigquery.BigNumericFieldType:
		return common.ValueTypeFloat
	case bigquery.BooleanFieldType:
		return common.ValueTypeBoolean
	case bigquery.TimestampFieldType:
		return common.ValueTypeDateTime
	case bigquery.DateFieldType:
		return common.ValueTypeDate
	case bigquery.TimeFieldType, bigquery.DateTimeFieldType:
		return common.ValueTypeDateTime
	case bigquery.GeographyFieldType:
		return common.ValueTypeOther
	case bigquery.RecordFieldType:
		return common.ValueTypeOther
	case bigquery.JSONFieldType:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}

// convertBigQueryRow converts a BigQuery row (map of Values) to a standard map.
func convertBigQueryRow(row map[string]bigquery.Value) map[string]any {
	result := make(map[string]any, len(row))

	for k, v := range row {
		result[k] = convertBigQueryValue(v)
	}

	return result
}

// convertBigQueryValue converts a BigQuery value to a Go native type.
func convertBigQueryValue(v bigquery.Value) any {
	if v == nil {
		return nil
	}

	switch typed := v.(type) {
	case []bigquery.Value:
		// Array type - convert each element
		result := make([]any, len(typed))
		for i, elem := range typed {
			result[i] = convertBigQueryValue(elem)
		}

		return result
	case []byte:
		return string(typed)
	case time.Time:
		return typed.Format(time.RFC3339)
	default:
		// For civil.Date, civil.Time, civil.DateTime, *big.Rat and other types
		// that implement fmt.Stringer, use their string representation.
		if stringer, ok := typed.(fmt.Stringer); ok {
			return stringer.String()
		}

		return typed
	}
}

// getFullyQualifiedTableName returns the fully qualified table name.
// Format: `project.dataset.table`
func (c *Connector) getFullyQualifiedTableName(tableName string) string {
	return fmt.Sprintf("`%s.%s.%s`",
		c.project,
		c.dataset,
		tableName,
	)
}

func (c *Connector) tablePath(table string) string {
	return fmt.Sprintf("projects/%s/datasets/%s/tables/%s",
		c.project,
		c.dataset,
		table,
	)
}

// boolPtr returns a pointer to a bool.
func boolPtr(b bool) *bool {
	return &b
}

// buildSetClause builds a SET clause for UPDATE statements.
func buildSetClause(record map[string]any) string {
	parts := make([]string, 0, len(record))

	for key := range record {
		parts = append(parts, fmt.Sprintf("%s = @%s", key, key))
	}

	return strings.Join(parts, ", ")
}

// buildQueryParameters converts a record map to BigQuery query parameters.
func buildQueryParameters(record map[string]any) []bigquery.QueryParameter {
	params := make([]bigquery.QueryParameter, 0, len(record))

	for key, value := range record {
		params = append(params, bigquery.QueryParameter{
			Name:  key,
			Value: value,
		})
	}

	return params
}
