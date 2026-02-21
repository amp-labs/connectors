package bigquery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	bqstorage "cloud.google.com/go/bigquery/storage/apiv1"
	"cloud.google.com/go/bigquery/storage/apiv1/storagepb"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"google.golang.org/api/option"
)

// =============================================================================
// Read performance & pagination
// =============================================================================
//
// These values were determined through benchmarking on a dataset with 141M rows.
// The goal was to find the sweet spot where each page fetch completes in ~30 seconds.
// At max 7 fields & 50K pageSize, we get a decent throughput while still allowing
// reasonable progress feedback. For a 141M row table, this means ~2,800 pages and
// ~23 hours total sync time. Dropping the field count & increasing the pageSize
// can improve this by a lot, but it will have to tweaked per builder.

const DefaultPageSize = 50000

// DefaultStreamCount controls how many parallel streams we use for reading.
//
// The Storage API can split a table into multiple streams that can be read
// concurrently. More streams = more parallelism = faster reads, but also
// more memory usage and connection overhead. It also means inexact page
// sizes.
const DefaultStreamCount = 4

// MaxFields is the hard limit on fields per read request.
// BigQuery is a columnar database and is fundamentally different
// from row-oriented databases where reading extra columns is nearly free.
// Reading columns is expensive and testing showed that more than 7 columns
// makes the connector too slow for usability. This may need review in future.
const MaxFields = 7

// Pagination:
// The Storage API uses server-side sessions and streams. We encode all the
// state needed to resume reading into an opaque token that gets passed back
// to the caller as NextPage.
//
// Token structure:
//   - SessionName: BigQuery Storage session ID (server-managed)
//   - Streams: Array of stream states, each tracking its own offset
//   - ActiveStreams: Count of streams not yet exhausted
// =============================================================================

// # Why Storage API instead of SQL?
// 1. Better pagination.
// 2. Parallel streams.
// 3. Arrow format.
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

	// Parse or create session token.
	var token *readSessionToken
	if params.NextPage != "" {
		token, err = parseReadSessionToken(string(params.NextPage))
		if err != nil {
			return nil, fmt.Errorf("invalid NextPage token: %w", err)
		}
	}

	// If no existing session, create a new read session.
	if token == nil {
		token, err = c.createReadSession(ctx, storageClient, params)
		if err != nil {
			return nil, fmt.Errorf("failed to create read session: %w", err)
		}
	}

	// Read rows from the current stream.
	rows, nextToken, done, err := c.readFromStream(ctx, storageClient, token, pageSize, params.Fields.List())
	if err != nil {
		return nil, err
	}

	var nextPage common.NextPageToken
	if !done && nextToken != nil {
		encoded, err := encodeReadSessionToken(nextToken)
		if err != nil {
			return nil, fmt.Errorf("failed to encode next page token: %w", err)
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

// The Storage API requires separate authentication from the BigQuery client.
// We use the credentials stored in the connector (passed via metadata).
func (c *Connector) getStorageClient(ctx context.Context) (*bqstorage.BigQueryReadClient, error) {
	return bqstorage.NewBigQueryReadClient(ctx,
		option.WithCredentialsJSON(c.credentials),
	)
}

// createReadSession creates a new BigQuery Storage read session.
func (c *Connector) createReadSession(
	ctx context.Context,
	client *bqstorage.BigQueryReadClient,
	params common.ReadParams,
) (*readSessionToken, error) {
	// Build selected fields if specified.
	var selectedFields []string
	if fields := params.Fields.List(); len(fields) > 0 {
		selectedFields = fields
	}

	// Create the read session request.
	// Using SelectedFields to only fetch requested columns for better performance.
	// Note: This means Raw will only contain the requested fields, not all fields.
	req := &storagepb.CreateReadSessionRequest{
		Parent: fmt.Sprintf("projects/%s", c.project),
		ReadSession: &storagepb.ReadSession{
			Table:      c.tablePath(params.ObjectName),
			DataFormat: storagepb.DataFormat_ARROW,
			ReadOptions: &storagepb.ReadSession_TableReadOptions{
				SelectedFields: selectedFields,
			},
		},

		MaxStreamCount: int32(DefaultStreamCount), // Use multiple streams for parallel reads.
	}

	// Add row restriction (filter) if specified.
	if params.Filter != "" {
		req.ReadSession.ReadOptions.RowRestriction = params.Filter
	}

	session, err := client.CreateReadSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create read session: %w", err)
	}

	if len(session.Streams) == 0 {
		// No data to read - return token with empty streams.
		return &readSessionToken{
			SessionName:   session.Name,
			Streams:       nil,
			ActiveStreams: 0,
		}, nil
	}

	// Build stream states for all streams.
	// Each stream starts at offset 0 and is marked as not done.
	streams := make([]streamState, len(session.Streams))
	for i, s := range session.Streams {
		streams[i] = streamState{
			Name:   s.Name,
			Offset: 0,
			Done:   false,
		}
	}

	return &readSessionToken{
		SessionName:   session.Name,
		Streams:       streams,
		ActiveStreams: len(streams),
	}, nil
}

// readFromStream reads rows from BigQuery Storage streams in parallel.
// Always uses multi-stream parallel reading for optimal performance.
func (c *Connector) readFromStream(
	ctx context.Context,
	client *bqstorage.BigQueryReadClient,
	token *readSessionToken,
	pageSize int,
	requestedFields []string,
) ([]common.ReadResultRow, *readSessionToken, bool, error) {
	// Check if no streams available.
	if len(token.Streams) == 0 || token.ActiveStreams == 0 {
		return nil, nil, true, nil
	}

	return c.readFromMultipleStreams(ctx, client, token, pageSize, requestedFields)
}

// streamResult holds the result from reading a single stream.
// Used to collect results from parallel stream reads.
type streamResult struct {
	StreamIndex int
	Rows        []common.ReadResultRow
	RowsRead    int64
	Done        bool
	Err         error
}

// readFromMultipleStreams reads rows from multiple streams in parallel.
//
// # How parallel streaming works
//
// BigQuery Storage API splits the table into multiple "streams" that can be
// read independently. We use the simultaneously package to read from all
// active streams concurrently, then combine the results.
//
// Example with 4 streams reading 50K rows:
//   - Each stream reads ~12.5K rows (50K / 4)
//   - All 4 reads happen in parallel via simultaneously.DoCtx
//   - Results are combined into a single page
//
// # Why we don't truncate results
//
// Each stream reads in Arrow batches (typically 1K-10K rows per batch).
// We can't stop mid-batch, so actual page sizes may exceed the requested
// size by up to one batch per stream. This is acceptable because:
//   - Data integrity is preserved (no row loss)
//   - Variance is typically <20% of page size
//   - Offset tracking remains accurate per-stream
func (c *Connector) readFromMultipleStreams(
	ctx context.Context,
	client *bqstorage.BigQueryReadClient,
	token *readSessionToken,
	pageSize int,
	requestedFields []string,
) ([]common.ReadResultRow, *readSessionToken, bool, error) {
	// Count active (non-done) streams.
	activeStreams := make([]int, 0, len(token.Streams))
	for i, s := range token.Streams {
		if !s.Done {
			activeStreams = append(activeStreams, i)
		}
	}

	if len(activeStreams) == 0 {
		return nil, nil, true, nil
	}

	// Calculate rows per stream (distribute evenly).
	rowsPerStream := (pageSize + len(activeStreams) - 1) / len(activeStreams)

	// Prepare result storage for each stream.
	// Each stream writes to its own slot, so no mutex needed.
	results := make([]streamResult, len(activeStreams))
	allocator := memory.NewGoAllocator()

	// Build jobs for the simultaneously package.
	// Each job reads from one stream and stores the result in its designated slot.
	jobs := make([]simultaneously.Job, len(activeStreams))
	for i, streamIdx := range activeStreams {
		// Capture loop variables for closure.
		resultIdx := i
		idx := streamIdx
		stream := token.Streams[idx]

		jobs[i] = func(ctx context.Context) error {
			result := c.readStreamChunk(ctx, client, stream, idx, rowsPerStream, requestedFields, allocator)
			results[resultIdx] = result

			if result.Err != nil {
				return result.Err
			}

			return nil
		}
	}

	// Execute all stream reads in parallel using simultaneously.
	// The package handles goroutine management, panic recovery, and cancellation.
	err := simultaneously.DoCtx(ctx, len(activeStreams), jobs...)
	if err != nil {
		return nil, nil, false, err
	}

	// Collect results from all streams.
	var allRows []common.ReadResultRow
	newStreams := make([]streamState, len(token.Streams))
	copy(newStreams, token.Streams)

	allDone := true

	for _, result := range results {
		allRows = append(allRows, result.Rows...)
		newStreams[result.StreamIndex].Offset += result.RowsRead
		newStreams[result.StreamIndex].Done = result.Done

		if !result.Done {
			allDone = false
		}
	}

	// Count remaining active streams.
	activeCount := 0
	for _, s := range newStreams {
		if !s.Done {
			activeCount++
		}
	}

	nextToken := &readSessionToken{
		SessionName:   token.SessionName,
		Streams:       newStreams,
		ActiveStreams: activeCount,
	}

	return allRows, nextToken, allDone, nil
}

// readStreamChunk reads a chunk of rows from a single stream.
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

	req := &storagepb.ReadRowsRequest{
		ReadStream: stream.Name,
		Offset:     stream.Offset,
	}

	rowStream, err := client.ReadRows(ctx, req)
	if err != nil {
		return streamResult{StreamIndex: streamIndex, Err: fmt.Errorf("failed to read rows from stream %d: %w", streamIndex, err)}
	}

	var rows []common.ReadResultRow
	var rowsRead int64
	var schemaBytes []byte

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

// parseArrowBatch parses an Arrow record batch into rows.
func parseArrowBatch(
	allocator memory.Allocator,
	schemaBytes []byte,
	batchBytes []byte,
	requestedFields []string,
) ([]common.ReadResultRow, error) {
	if len(batchBytes) == 0 {
		return nil, nil
	}

	// First, read the schema from the schema bytes.
	schemaReader := ipc.NewMessageReader(bytes.NewReader(schemaBytes), ipc.WithAllocator(allocator))
	schemaMsg, err := schemaReader.Message()
	if err != nil {
		return nil, fmt.Errorf("failed to read schema message: %w", err)
	}
	defer schemaMsg.Release()

	if schemaMsg.Type() != ipc.MessageSchema {
		return nil, fmt.Errorf("expected schema message, got %v", schemaMsg.Type())
	}

	// Create a reader that combines schema and batch using NewReaderFromMessageReader.
	// The schema message is already read, so we create a combined reader.
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

// getArrowValue extracts a value from an Arrow array at the given index.
//
// # Arrow type mapping
//
// BigQuery types are serialized as Arrow types, which we convert to Go types:
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
// # Why strings for dates/timestamps?
//
// We serialize temporal types as strings for JSON compatibility and to avoid
// timezone confusion. The formats are ISO 8601 compliant.
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
		// Handle array/repeated fields.
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
		// Handle struct/record fields.
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
		// For complex types, return string representation.
		return fmt.Sprintf("%v", arr.ValueStr(idx))
	}
}

// arrowBatchReader implements io.Reader for Arrow IPC format.
type arrowBatchReader struct {
	data   []byte
	offset int
}

func (r *arrowBatchReader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

// parseReadSessionToken decodes a NextPage token.
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

// encodeReadSessionToken encodes a token for use as NextPage.
func encodeReadSessionToken(token *readSessionToken) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
