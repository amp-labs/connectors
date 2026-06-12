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

// Fathom has conservative rate limits for the recordings API.
// https://developers.fathom.ai/api-overview#heavy-requests-rate-limits
const maxConcurrentMeetingRecordingFetch = 4

func (c *Connector) enrichMeetingsWithRecordings(
	ctx context.Context,
	rows []common.ReadResultRow,
	fields datautils.StringSet,
) error {
	needsSummary := fields.Has("default_summary")
	needsTranscript := fields.Has("transcript")

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

				rows[idx].Fields["default_summary"] = summary
				rows[idx].Raw["default_summary"] = summary
			}

			if needsTranscript {
				transcript, err := c.fetchRecordingTranscript(ctx, recordingID)
				if err != nil {
					return fmt.Errorf("fetching transcript for recording %s: %w", recordingID, err)
				}

				rows[idx].Fields["transcript"] = transcript
				rows[idx].Raw["transcript"] = transcript
			}

			return nil
		}
	}

	return simultaneously.DoCtx(ctx, maxConcurrentMeetingRecordingFetch, jobs...)
}

func recordingIDFromRaw(raw map[string]any) (string, error) {
	value, ok := raw["recording_id"]
	if !ok || value == nil {
		return "", fmt.Errorf("%w: recording_id", common.ErrMissingExpectedValues)
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
		return "", fmt.Errorf("unexpected recording_id type %T", value)
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
		return nil, nil
	}

	summary, ok := (*body)["summary"]
	if !ok || summary == nil {
		return nil, nil
	}

	summaryMap, ok := summary.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected summary type %T", summary)
	}

	return summaryMap, nil
}

// fetchRecordingTranscript fetches the transcript for a given recording ID.
// https://developers.fathom.ai/api-reference/recordings/get-transcript
func (c *Connector) fetchRecordingTranscript(ctx context.Context, recordingID string) (any, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, apiVersion, "recordings", recordingID, "transcript")
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
		return nil, nil
	}

	transcript, ok := (*body)["transcript"]
	if !ok {
		return nil, nil
	}

	return transcript, nil
}
