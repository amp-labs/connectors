package fathom

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

const (
	fieldDefaultSummary = "default_summary"
	fieldTranscript     = "transcript"
	fieldRecordingID    = "recording_id"
)

// Fathom has conservative rate limits for the recordings API.
// According to https://developers.fathom.ai/api-overview#heavy-requests-rate-limits
// "During periods of elevated activity this limit may be adjusted down to 5 every 60 seconds."
// Each job is potentially 2 API calls (transcript and summary),
// so we cap concurrency at 2 (4 API calls) and wait
// 60 second between batches to avoid hitting the rate limit.
const maxConcurrentMeetingRecordingFetch = 2
const waitIntervalMeetingRecordingFetchMS = 60 * 1000 // 60 seconds

// enrichMeetingsWithRecordings fetches default_summary and/or transcript for each
// meeting row from the recordings API. OAuth-connected apps cannot use
// include_summary or include_transcript on the list-meetings endpoint, so these
// fields are loaded per recording instead.
//
// Rows are mutated in place: each row's Fields and Raw maps are updated with the
// fetched values. Fetches run concurrently, capped at maxConcurrentMeetingRecordingFetch.
func (c *Connector) enrichMeetingsWithRecordings(
	ctx context.Context,
	rows []common.ReadResultRow,
	fields datautils.StringSet,
) error {
	needsSummary := fields.Has(fieldDefaultSummary)
	needsTranscript := fields.Has(fieldTranscript)

	if !needsSummary && !needsTranscript {
		return nil
	}

	jobs := make([]simultaneously.Job, len(rows))

	for i := range rows {
		idx := i

		recordingID, err := recordingIDFromRaw(rows[i].Raw)
		if err != nil {
			return fmt.Errorf("enriching meeting at index %d: %w", idx, err)
		}

		jobs[idx] = func(ctx context.Context) error {
			if needsSummary {
				summary, err := c.fetchRecordingSummary(ctx, recordingID)
				if err != nil {
					return fmt.Errorf("fetching summary for recording %s: %w", recordingID, err)
				}

				rows[idx].Fields[fieldDefaultSummary] = summary
				rows[idx].Raw[fieldDefaultSummary] = summary
			}

			if needsTranscript {
				transcript, err := c.fetchRecordingTranscript(ctx, recordingID)
				if err != nil {
					return fmt.Errorf("fetching transcript for recording %s: %w", recordingID, err)
				}

				rows[idx].Fields[fieldTranscript] = transcript
				rows[idx].Raw[fieldTranscript] = transcript
			}

			return nil
		}
	}

	return simultaneously.DoCtxWithWaitInterval(ctx,
		maxConcurrentMeetingRecordingFetch, waitIntervalMeetingRecordingFetchMS, jobs...)
}

func recordingIDFromRaw(raw map[string]any) (string, error) {
	value, ok := raw[fieldRecordingID]
	if !ok || value == nil {
		return "", fmt.Errorf("%w: %s", common.ErrMissingExpectedValues, fieldRecordingID)
	}

	switch id := value.(type) {
	case float64:
		return strconv.FormatFloat(id, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(id), nil
	case int64:
		return strconv.FormatInt(id, 10), nil
	case json.Number:
		return id.String(), nil
	default:
		return "", fmt.Errorf("%w: %T", ErrUnexpectedRecordingIDType, value)
	}
}

// fetchRecordingSummary fetches the summary for a given recording ID.
// https://developers.fathom.ai/api-reference/recordings/get-summary
func (c *Connector) fetchRecordingSummary(ctx context.Context, recordingID string) (map[string]any, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, apiVersion, "recordings", recordingID, "summary")
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, err
	}

	if body == nil {
		return map[string]any{}, nil
	}

	summary, ok := (*body)["summary"]
	if !ok || summary == nil {
		return map[string]any{}, nil
	}

	summaryMap, ok := summary.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: %T", ErrUnexpectedSummaryType, summary)
	}

	return summaryMap, nil
}

// fetchRecordingTranscript fetches the transcript for a given recording ID.
// https://developers.fathom.ai/api-reference/recordings/get-transcript
func (c *Connector) fetchRecordingTranscript(ctx context.Context, recordingID string) (any, error) {
	url, err := urlbuilder.New(
		c.ProviderInfo().BaseURL, restAPIPrefix, apiVersion, "recordings", recordingID, "transcript",
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, err
	}

	if body == nil {
		return []any{}, nil
	}

	transcript, ok := (*body)["transcript"]
	if !ok || transcript == nil {
		return []any{}, nil
	}

	return transcript, nil
}
