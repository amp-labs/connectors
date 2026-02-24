package bigquery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	bqstorage "cloud.google.com/go/bigquery/storage/apiv1"
	"cloud.google.com/go/bigquery/storage/apiv1/storagepb"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// =============================================================================
// Read strategy overview
// =============================================================================
//
// The connector supports two read modes, determined by whether Since is set:
//
//  1. INCREMENTAL READ (Since and/or Until set by server)
//     The server provides a narrow time range (typically minutes).
//     We create a single Storage API session with a RowRestriction:
//       timestampColumn >= Since AND timestampColumn < Until
//     The data volume is small, so the session completes in one or two pages.
//     No windowing needed.
//
//  2. BACKFILL (no Since, no NextPage — first read ever)
//     The table may contain billions of rows spanning years. A single Storage API
//     session cannot handle this because sessions expire after 6 hours.
//
//     Solution: time-based windowing.
//     The connector partitions the table into 30-day windows and reads one window
//     at a time. Each window gets its own Storage API session. When a window is
//     exhausted, the connector advances to the next window. Empty windows (no data
//     in that time range) are skipped instantly.
//
//     Example for a 1B-row table spanning 2019–2024:
//       Window 1: [2000-01-01, 2000-01-31) → 0 rows, skip instantly
//       ...
//       Window 228: [2019-01-01, 2019-01-31) → 14M rows → 280 pages → ~2.3 hours
//       Window 229: [2019-01-31, 2019-03-02) → 14M rows → 280 pages → ~2.3 hours
//       ...
//       Window 300: [2025-01-01, 2025-01-31) → 0 rows → backfill complete
//
//     Each window's session lives for ~2.3 hours, well within the 6-hour limit.
//
// =============================================================================
//
// Session expiry handling
// =============================================================================
//
// If a Storage API session expires mid-read (after 6 hours), the gRPC layer
// returns FAILED_PRECONDITION with message "session expired". When detected:
//   - We create a NEW session for the SAME window (same RowRestriction)
//   - We re-read from the beginning of that window
//   - The Ampersand server deduplicates by primary key, so re-sent rows are safe
//   - We return the new session's token so subsequent pages use the fresh session
//
// This is acceptable because:
//   - Each window is ~30 days, so sessions rarely expire (~2.3h for dense data)
//   - Re-reading a window wastes at most one window's worth of work
//   - The server's deduplication guarantees correctness
//
// =============================================================================

// Performance tuning constants.
// Determined through benchmarking on a 141M-row dataset.
const (
	// DefaultPageSize is the target number of rows per Read() call.
	// At 7 fields and 50K rows, each page completes in ~30 seconds.
	DefaultPageSize = 50000

	// DefaultStreamCount controls parallel stream count for the Storage API.
	// More streams = more parallelism but also more memory and connection overhead.
	// 4 streams provide a good balance for the 50K page size.
	DefaultStreamCount = 4

	// MaxFields caps the number of columns per read.
	// BigQuery is columnar — reading extra columns is expensive, unlike row-oriented DBs.
	// Benchmarking showed >7 columns makes reads too slow for usability.
	MaxFields = 7
)

// Read fetches rows from a BigQuery table using the Storage Read API.
//
// The method is stateless: all state needed to resume is encoded in the NextPage token.
// The Ampersand server calls Read() in a loop, passing the returned NextPage token
// back on each call, until Done=true.
//
// First call (no NextPage):
//   - If Since is set → incremental read with RowRestriction
//   - If Since is zero → backfill with automatic time windowing
//
// Subsequent calls (NextPage set):
//   - Resume from the encoded session/window state
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if err := validateFieldCount(params.Fields); err != nil {
		return nil, err
	}

	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	storageClient, err := c.getStorageClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}
	defer storageClient.Close()

	// Determine the read token: either parse from NextPage or create initial state.
	token, err := c.resolveToken(params)
	if err != nil {
		return nil, err
	}

	// Main read loop. May iterate multiple times if we hit empty windows during backfill.
	for {
		// Ensure we have an active session for the current window.
		if token.SessionName == "" {
			token, err = c.createSessionForToken(ctx, storageClient, token, params)
			if err != nil {
				return nil, err
			}

			// Empty window: no data matched the RowRestriction.
			if len(token.Streams) == 0 || token.ActiveStreams == 0 {
				if token.IsBackfill {
					// Advance to next window and retry.
					token, err = c.advanceWindow(token)
					if err != nil {
						// advanceWindow returns nil token when all windows exhausted.
						return &common.ReadResult{Done: true}, nil
					}

					continue // Try next window.
				}

				// Incremental read with no matching rows.
				return &common.ReadResult{Done: true}, nil
			}
		}

		// Read a page of rows from the current session.
		rows, nextToken, done, err := c.readFromStreams(ctx, storageClient, token, pageSize, params)
		if err != nil {
			// Check for session expiry. If expired, create a new session for the
			// same window and restart from the beginning of the window.
			if isSessionExpired(err) {
				token.SessionName = ""
				token.Streams = nil
				token.ActiveStreams = 0

				continue // Will create a new session on next iteration.
			}

			return nil, err
		}

		// Window exhausted?
		if done && nextToken != nil && nextToken.IsBackfill {
			// Try advancing to the next window.
			advanced, advErr := c.advanceWindow(nextToken)
			if advErr != nil {
				// All windows exhausted — the entire backfill is complete.
				return &common.ReadResult{
					Rows: int64(len(rows)),
					Data: rows,
					Done: true,
				}, nil
			}

			// More windows to go. Encode the advanced token.
			encoded, encErr := encodeReadSessionToken(advanced)
			if encErr != nil {
				return nil, fmt.Errorf("failed to encode next page token: %w", encErr)
			}

			return &common.ReadResult{
				Rows:     int64(len(rows)),
				Data:     rows,
				NextPage: common.NextPageToken(encoded),
				Done:     false,
			}, nil
		}

		// Normal case: more data in current session, or incremental read done.
		var nextPage common.NextPageToken
		if !done && nextToken != nil {
			encoded, encErr := encodeReadSessionToken(nextToken)
			if encErr != nil {
				return nil, fmt.Errorf("failed to encode next page token: %w", encErr)
			}

			nextPage = common.NextPageToken(encoded)
		}

		return &common.ReadResult{
			Rows:     int64(len(rows)),
			Data:     rows,
			NextPage: nextPage,
			Done:     done,
		}, nil
	}
}

// =============================================================================
// Token resolution
// =============================================================================

// resolveToken determines the initial read token from the ReadParams.
//
// Three cases:
//  1. NextPage is set → parse the token (resume a previous read)
//  2. Since is set → incremental read (no windowing, just RowRestriction)
//  3. Neither → backfill (start windowing from epoch)
func (c *Connector) resolveToken(params common.ReadParams) (*readSessionToken, error) {
	// Case 1: Resuming from a previous page.
	if params.NextPage != "" {
		token, err := parseReadSessionToken(string(params.NextPage))
		if err != nil {
			return nil, fmt.Errorf("invalid NextPage token: %w", err)
		}

		return token, nil
	}

	// Case 2: Incremental read (Since set by server).
	if !params.Since.IsZero() {
		return &readSessionToken{
			IsBackfill: false,
			// No window boundaries — the RowRestriction is built from params.Since/Until directly.
		}, nil
	}

	// Case 3: Backfill — start windowing from epoch.
	now := time.Now().UTC()
	windowEnd := backfillEpoch.Add(backfillWindowSize)

	if windowEnd.After(now) {
		windowEnd = now
	}

	return &readSessionToken{
		IsBackfill:  true,
		WindowStart: backfillEpoch.Format(time.RFC3339),
		WindowEnd:   windowEnd.Format(time.RFC3339),
	}, nil
}

// =============================================================================
// Session creation
// =============================================================================

// createSessionForToken creates a new Storage API session for the given token.
// The RowRestriction is built from the token's window boundaries (backfill)
// or from params.Since/Until (incremental).
func (c *Connector) createSessionForToken(
	ctx context.Context,
	client *bqstorage.BigQueryReadClient,
	token *readSessionToken,
	params common.ReadParams,
) (*readSessionToken, error) {
	rowRestriction := c.buildRowRestriction(token, params)

	var selectedFields []string
	if fields := params.Fields.List(); len(fields) > 0 {
		selectedFields = fields
	}

	req := &storagepb.CreateReadSessionRequest{
		Parent: fmt.Sprintf("projects/%s", c.project),
		ReadSession: &storagepb.ReadSession{
			Table:      c.tablePath(params.ObjectName),
			DataFormat: storagepb.DataFormat_ARROW,
			ReadOptions: &storagepb.ReadSession_TableReadOptions{
				SelectedFields: selectedFields,
				RowRestriction: rowRestriction,
			},
		},
		MaxStreamCount: int32(DefaultStreamCount),
	}

	session, err := client.CreateReadSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create read session: %w", err)
	}

	// Preserve window state from the input token.
	newToken := &readSessionToken{
		SessionName: session.Name,
		IsBackfill:  token.IsBackfill,
		WindowStart: token.WindowStart,
		WindowEnd:   token.WindowEnd,
	}

	if len(session.Streams) == 0 {
		newToken.Streams = nil
		newToken.ActiveStreams = 0

		return newToken, nil
	}

	streams := make([]streamState, len(session.Streams))
	for i, s := range session.Streams {
		streams[i] = streamState{
			Name:   s.Name,
			Offset: 0,
			Done:   false,
		}
	}

	newToken.Streams = streams
	newToken.ActiveStreams = len(streams)

	return newToken, nil
}

// buildRowRestriction constructs the WHERE clause for the Storage API session.
//
// For incremental reads: timestampColumn >= Since AND timestampColumn < Until
// For backfill windows:  timestampColumn >= WindowStart AND timestampColumn < WindowEnd
//
// The RowRestriction is applied server-side before stream creation, so only
// matching rows are included in the session. This is what makes windowing work:
// each window only contains its slice of the data.
func (c *Connector) buildRowRestriction(token *readSessionToken, params common.ReadParams) string {
	var conditions []string

	if token.IsBackfill {
		// Backfill: use window boundaries from the token.
		if token.WindowStart != "" {
			conditions = append(conditions,
				fmt.Sprintf("%s >= TIMESTAMP('%s')", c.timestampColumn, token.WindowStart))
		}

		if token.WindowEnd != "" {
			conditions = append(conditions,
				fmt.Sprintf("%s < TIMESTAMP('%s')", c.timestampColumn, token.WindowEnd))
		}
	} else {
		// Incremental: use Since/Until from params.
		if !params.Since.IsZero() {
			conditions = append(conditions,
				fmt.Sprintf("%s >= TIMESTAMP('%s')", c.timestampColumn, params.Since.UTC().Format(time.RFC3339)))
		}

		if !params.Until.IsZero() {
			conditions = append(conditions,
				fmt.Sprintf("%s < TIMESTAMP('%s')", c.timestampColumn, params.Until.UTC().Format(time.RFC3339)))
		}
	}

	// Append any user-provided filter.
	if params.Filter != "" {
		conditions = append(conditions, params.Filter)
	}

	return strings.Join(conditions, " AND ")
}

// =============================================================================
// Window management
// =============================================================================

// advanceWindow moves to the next 30-day window for backfill.
// Returns an error if all windows are exhausted (WindowEnd > now).
func (c *Connector) advanceWindow(token *readSessionToken) (*readSessionToken, error) {
	windowEnd, err := time.Parse(time.RFC3339, token.WindowEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid WindowEnd in token: %w", err)
	}

	now := time.Now().UTC()
	if windowEnd.After(now) || windowEnd.Equal(now) {
		// We've reached the present — backfill is complete.
		return nil, fmt.Errorf("backfill complete: window end %s >= now", windowEnd.Format(time.RFC3339))
	}

	newStart := windowEnd
	newEnd := newStart.Add(backfillWindowSize)

	if newEnd.After(now) {
		newEnd = now
	}

	return &readSessionToken{
		IsBackfill:  true,
		WindowStart: newStart.Format(time.RFC3339),
		WindowEnd:   newEnd.Format(time.RFC3339),
		// SessionName, Streams, ActiveStreams left empty — new session needed.
	}, nil
}

// =============================================================================
// Session expiry detection
// =============================================================================

// isSessionExpired checks if an error indicates that the Storage API session
// has expired. The BigQuery Storage API returns:
//
//	gRPC code: FAILED_PRECONDITION (9)
//	Message:   "... session expired"
//
// Sessions expire 6 hours after creation. There is no API to extend them.
// When detected, the caller should create a new session for the same window.
func isSessionExpired(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	return st.Code() == codes.FailedPrecondition &&
		strings.Contains(st.Message(), "session expired")
}

// =============================================================================
// Storage client
// =============================================================================

// getStorageClient creates a new BigQuery Storage Read API gRPC client.
// The Storage API requires separate authentication from the BigQuery SQL client
// because it uses gRPC transport, not HTTP.
func (c *Connector) getStorageClient(ctx context.Context) (*bqstorage.BigQueryReadClient, error) {
	return bqstorage.NewBigQueryReadClient(ctx,
		option.WithCredentialsJSON(c.credentials),
	)
}

// =============================================================================
// Parallel stream reading
// =============================================================================

// streamResult holds the result from reading a single stream.
type streamResult struct {
	StreamIndex int
	Rows        []common.ReadResultRow
	RowsRead    int64
	Done        bool
	Err         error
}

// readFromStreams reads rows from multiple Storage API streams in parallel.
//
// BigQuery splits each session into multiple streams that can be read concurrently.
// We use the simultaneously package to read from all active streams in parallel,
// then combine the results into a single page.
//
// Example with 4 streams and a 50K-row page:
//   - Each stream reads ~12.5K rows (50K / 4)
//   - All 4 reads happen concurrently via simultaneously.DoCtx
//   - Results are merged into one page of ~50K rows
//
// Actual page sizes may exceed the target by up to one Arrow batch per stream
// (~1K–10K rows) because we can't stop mid-batch without losing data.
func (c *Connector) readFromStreams(
	ctx context.Context,
	client *bqstorage.BigQueryReadClient,
	token *readSessionToken,
	pageSize int,
	params common.ReadParams,
) ([]common.ReadResultRow, *readSessionToken, bool, error) {
	if len(token.Streams) == 0 || token.ActiveStreams == 0 {
		return nil, token, true, nil
	}

	// Find active (non-done) streams.
	activeStreams := make([]int, 0, len(token.Streams))
	for i, s := range token.Streams {
		if !s.Done {
			activeStreams = append(activeStreams, i)
		}
	}

	if len(activeStreams) == 0 {
		return nil, token, true, nil
	}

	// Distribute rows evenly across active streams.
	rowsPerStream := (pageSize + len(activeStreams) - 1) / len(activeStreams)

	// Each stream writes to its own slot — no mutex needed.
	results := make([]streamResult, len(activeStreams))
	allocator := memory.NewGoAllocator()
	requestedFields := params.Fields.List()

	// Build parallel jobs.
	// Stream errors are stored in results, NOT returned to simultaneously.
	// This prevents one failed stream from cancelling the others.
	// Failed streams are retried on the next page (their offsets stay unchanged).
	jobs := make([]simultaneously.Job, len(activeStreams))
	for i, streamIdx := range activeStreams {
		resultIdx := i
		idx := streamIdx
		stream := token.Streams[idx]

		jobs[i] = func(ctx context.Context) error {
			results[resultIdx] = c.readStreamChunk(ctx, client, stream, idx, rowsPerStream, requestedFields, allocator)

			// Always return nil — errors are handled after all streams finish.
			return nil
		}
	}

	// Execute all stream reads in parallel.
	// The simultaneously package handles goroutine management and panic recovery.
	// Individual stream failures do NOT cancel other streams.
	if err := simultaneously.DoCtx(ctx, len(activeStreams), jobs...); err != nil {
		// This only triggers on panics, not on stream errors.
		return nil, nil, false, err
	}

	// Triage results: check for session expiry, count failures, collect rows.
	var (
		allRows     []common.ReadResultRow
		streamErrs  []error
		succeededAny bool
	)

	newStreams := make([]streamState, len(token.Streams))
	copy(newStreams, token.Streams)

	for _, result := range results {
		if result.Err != nil {
			// Session expiry must be handled at the session level — bubble it up
			// so Read() can reset the session for the entire window.
			if isSessionExpired(result.Err) {
				return nil, nil, false, result.Err
			}

			streamErrs = append(streamErrs, result.Err)
			// Leave this stream's offset unchanged — it will be retried next page.
			continue
		}

		succeededAny = true
		allRows = append(allRows, result.Rows...)
		newStreams[result.StreamIndex].Offset += result.RowsRead
		newStreams[result.StreamIndex].Done = result.Done
	}

	// If ALL streams failed, nothing was salvaged — return the errors.
	if !succeededAny {
		return nil, nil, false, fmt.Errorf("all streams failed: %w", errors.Join(streamErrs...))
	}

	allDone := true

	activeCount := 0
	for _, s := range newStreams {
		if !s.Done {
			activeCount++
			allDone = false
		}
	}

	nextToken := &readSessionToken{
		SessionName:  token.SessionName,
		Streams:      newStreams,
		ActiveStreams: activeCount,
		IsBackfill:   token.IsBackfill,
		WindowStart:  token.WindowStart,
		WindowEnd:    token.WindowEnd,
	}

	return allRows, nextToken, allDone, nil
}

// readStreamChunk reads a chunk of rows from a single stream.
// Called concurrently — one goroutine per active stream.
func (c *Connector) readStreamChunk(
	ctx context.Context,
	client *bqstorage.BigQueryReadClient,
	stream streamState,
	streamIndex int,
	maxRows int,
	requestedFields []string,
	allocator memory.Allocator,
) streamResult {
	if stream.Done {
		return streamResult{StreamIndex: streamIndex, Done: true}
	}

	rowStream, err := client.ReadRows(ctx, &storagepb.ReadRowsRequest{
		ReadStream: stream.Name,
		Offset:     stream.Offset,
	})
	if err != nil {
		return streamResult{StreamIndex: streamIndex, Err: fmt.Errorf("failed to read rows from stream %d: %w", streamIndex, err)}
	}

	var (
		rows        []common.ReadResultRow
		rowsRead    int64
		schemaBytes []byte
	)

	for len(rows) < maxRows {
		response, err := rowStream.Recv()
		if err == io.EOF {
			return streamResult{StreamIndex: streamIndex, Rows: rows, RowsRead: rowsRead, Done: true}
		}

		if err != nil {
			return streamResult{StreamIndex: streamIndex, Err: fmt.Errorf("failed to receive rows from stream %d: %w", streamIndex, err)}
		}

		if arrowSchema := response.GetArrowSchema(); arrowSchema != nil {
			if schema := arrowSchema.GetSerializedSchema(); len(schema) > 0 {
				schemaBytes = schema
			}
		}

		arrowData := response.GetArrowRecordBatch()
		if arrowData == nil {
			continue
		}

		if len(schemaBytes) == 0 {
			return streamResult{StreamIndex: streamIndex, Err: fmt.Errorf("no Arrow schema received before record batch in stream %d", streamIndex)}
		}

		batchRows, err := parseArrowBatch(allocator, schemaBytes, arrowData.SerializedRecordBatch, requestedFields)
		if err != nil {
			return streamResult{StreamIndex: streamIndex, Err: fmt.Errorf("failed to parse Arrow batch in stream %d: %w", streamIndex, err)}
		}

		rows = append(rows, batchRows...)
		rowsRead += int64(len(batchRows))

		if len(rows) >= maxRows {
			break
		}
	}

	return streamResult{StreamIndex: streamIndex, Rows: rows, RowsRead: rowsRead, Done: false}
}

// =============================================================================
// Arrow parsing
// =============================================================================

// parseArrowBatch parses an Arrow IPC record batch into ReadResultRows.
//
// The Storage API returns data as serialized Arrow record batches. Each batch
// contains a schema message followed by a record batch message. We combine them
// into a single reader and iterate over the rows.
func parseArrowBatch(
	allocator memory.Allocator,
	schemaBytes []byte,
	batchBytes []byte,
	requestedFields []string,
) ([]common.ReadResultRow, error) {
	if len(batchBytes) == 0 {
		return nil, nil
	}

	// Validate the schema message before combining.
	schemaReader := ipc.NewMessageReader(bytes.NewReader(schemaBytes), ipc.WithAllocator(allocator))

	schemaMsg, err := schemaReader.Message()
	if err != nil {
		return nil, fmt.Errorf("failed to read schema message: %w", err)
	}
	defer schemaMsg.Release()

	if schemaMsg.Type() != ipc.MessageSchema {
		return nil, fmt.Errorf("expected schema message, got %v", schemaMsg.Type())
	}

	// Combine schema + batch into a single IPC stream for the Arrow reader.
	combinedReader := io.MultiReader(
		bytes.NewReader(schemaBytes),
		bytes.NewReader(batchBytes),
	)

	reader, err := ipc.NewReader(combinedReader, ipc.WithAllocator(allocator))
	if err != nil {
		return nil, fmt.Errorf("failed to create Arrow reader: %w", err)
	}
	defer reader.Release()

	var rows []common.ReadResultRow

	for reader.Next() {
		record := reader.Record()
		schema := record.Schema()

		numRows := int(record.NumRows())
		numCols := int(record.NumCols())

		for i := 0; i < numRows; i++ {
			raw := make(map[string]any, numCols)

			for j := 0; j < numCols; j++ {
				field := schema.Field(j)
				col := record.Column(j)
				raw[field.Name] = getArrowValue(col, i)
			}

			fields := common.ExtractLowercaseFieldsFromRaw(requestedFields, raw)

			rows = append(rows, common.ReadResultRow{
				Fields: fields,
				Raw:    raw,
			})
		}
	}

	if err := reader.Err(); err != nil {
		return nil, fmt.Errorf("error reading Arrow data: %w", err)
	}

	return rows, nil
}

// getArrowValue extracts a Go value from an Arrow array at the given index.
//
// Arrow type mapping:
//
//	BigQuery Type    Arrow Type       Go Type
//	─────────────────────────────────────────────────────
//	INT64            Int64            int64
//	FLOAT64          Float64          float64
//	STRING           String           string
//	BOOL             Boolean          bool
//	BYTES            Binary           []byte
//	DATE             Date32/Date64    string ("2006-01-02")
//	TIMESTAMP        Timestamp        string (RFC3339)
//	RECORD/STRUCT    Struct           map[string]any
//	REPEATED/ARRAY   List             []any
//
// Temporal types are serialized as ISO 8601 strings for JSON compatibility.
func getArrowValue(arr arrow.Array, idx int) any {
	if arr.IsNull(idx) {
		return nil
	}

	switch typed := arr.(type) {
	case *array.Int64:
		return typed.Value(idx)
	case *array.Int32:
		return typed.Value(idx)
	case *array.Int16:
		return typed.Value(idx)
	case *array.Int8:
		return typed.Value(idx)
	case *array.Uint64:
		return typed.Value(idx)
	case *array.Uint32:
		return typed.Value(idx)
	case *array.Uint16:
		return typed.Value(idx)
	case *array.Uint8:
		return typed.Value(idx)
	case *array.Float64:
		return typed.Value(idx)
	case *array.Float32:
		return float64(typed.Value(idx))
	case *array.String:
		return typed.Value(idx)
	case *array.Boolean:
		return typed.Value(idx)
	case *array.Binary:
		return typed.Value(idx)
	case *array.Date32:
		return typed.Value(idx).ToTime().Format("2006-01-02")
	case *array.Date64:
		return typed.Value(idx).ToTime().Format("2006-01-02")
	case *array.Timestamp:
		return typed.Value(idx).ToTime(typed.DataType().(*arrow.TimestampType).Unit).Format("2006-01-02T15:04:05Z07:00")
	case *array.List:
		if typed.IsNull(idx) {
			return nil
		}

		start, end := typed.ValueOffsets(idx)
		values := typed.ListValues()
		result := make([]any, 0, end-start)

		for i := int(start); i < int(end); i++ {
			result = append(result, getArrowValue(values, i))
		}

		return result
	case *array.Struct:
		if typed.IsNull(idx) {
			return nil
		}

		structType := typed.DataType().(*arrow.StructType)
		result := make(map[string]any, structType.NumFields())

		for i := 0; i < structType.NumFields(); i++ {
			field := structType.Field(i)
			result[field.Name] = getArrowValue(typed.Field(i), idx)
		}

		return result
	default:
		return fmt.Sprintf("%v", arr.ValueStr(idx))
	}
}

// =============================================================================
// Token serialization
// =============================================================================

func parseReadSessionToken(encoded string) (*readSessionToken, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var token readSessionToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func encodeReadSessionToken(token *readSessionToken) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
