package bigquery

import (
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// =============================================================================
// Pagination token types
// =============================================================================
//
// The NextPage token encodes two layers of state:
//
//  1. Session state (within a window): which Storage API session we're reading,
//     which streams exist, and how far into each stream we've read.
//
//  2. Window state (across windows): the current time window boundaries for
//     backfill windowing. This allows the connector to partition a large table
//     into manageable chunks that each fit within the Storage API's 6-hour
//     session lifetime.
//
// During a backfill, the flow is:
//
//	Window 1: [epoch, epoch+30d) → create session → read all pages → window done
//	Window 2: [epoch+30d, epoch+60d) → create session → read all pages → window done
//	...
//	Window N: [T, T+30d) → empty result → advance → all windows done → Done=true
//
// During incremental reads (Since/Until set by server):
//
//	Single session with RowRestriction on timestampColumn → read all pages → Done=true
//	No windowing needed because the data volume is small.
//
// If a session expires mid-window (gRPC FAILED_PRECONDITION "session expired"):
//
//	The connector creates a new session for the SAME window and re-reads from
//	the beginning of that window. The server deduplicates by primary key.
//	Each window is ~30 days of data, so sessions typically complete in minutes,
//	making expiry unlikely but handled gracefully.

// readSessionToken holds all state needed to resume a paginated read.
// Serialized to JSON, base64-encoded, and returned as NextPage.
type readSessionToken struct {
	// --- Session state (within a window) ---

	// SessionName is the BigQuery Storage session identifier.
	// Sessions are server-managed and expire after 6 hours.
	// Empty string means a new session needs to be created.
	SessionName string `json:"session,omitempty"`

	// Streams contains the state of each parallel stream in the current session.
	// Each stream tracks its own offset and completion state independently.
	Streams []streamState `json:"streams,omitempty"`

	// ActiveStreams is the count of streams that still have data to read.
	// When this reaches 0, the current window is exhausted.
	ActiveStreams int `json:"activeStreams"`

	// --- Window state (backfill only) ---

	// WindowStart is the inclusive lower bound of the current time window (RFC3339).
	// For backfills, this advances by WindowSize after each window is exhausted.
	// Empty for incremental reads (Since/Until handles the filtering).
	WindowStart string `json:"windowStart,omitempty"`

	// WindowEnd is the exclusive upper bound of the current time window (RFC3339).
	// The RowRestriction uses: timestampColumn >= WindowStart AND timestampColumn < WindowEnd.
	WindowEnd string `json:"windowEnd,omitempty"`

	// IsBackfill indicates this token is part of a windowed backfill.
	// When true, exhausting a window advances to the next window instead of returning Done.
	// When false, exhausting the session means Done=true.
	IsBackfill bool `json:"isBackfill,omitempty"`
}

// streamState tracks the read position within a single BigQuery Storage stream.
type streamState struct {
	// Name is the fully-qualified stream resource name assigned by BigQuery.
	Name string `json:"name"`

	// Offset is the number of rows already read from this stream.
	// Used to resume reading from the correct position within a session.
	Offset int64 `json:"offset"`

	// Done indicates this stream has been fully consumed (EOF reached).
	Done bool `json:"done"`
}

// backfillWindowSize is the duration of each time window during a backfill.
//
// Why 30 days: Storage API sessions expire after 6 hours. A 30-day window of a
// billion-row table spanning 6 years contains ~14M rows (~280 pages at 50K/page),
// which takes ~2.3 hours — well within the 6-hour limit. Smaller windows would
// work too but create more sessions (more overhead). Larger windows risk session
// expiry on very dense tables.
const backfillWindowSize = 30 * 24 * time.Hour

// backfillEpoch is the starting point for backfill windowing when no Since is provided.
// We use 2000-01-01 as a reasonable lower bound — BigQuery was launched in 2010,
// and most data timestamps post-date this. Windows before the actual data range
// return 0 rows and are skipped instantly.
var backfillEpoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

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
