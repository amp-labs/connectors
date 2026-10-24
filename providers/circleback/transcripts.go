package circleback

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/amp-labs/amp-common/simultaneously"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

const (
	// fieldTranscript is the meetings field holding the transcript, which the list does not return.
	fieldTranscript = "transcript"

	// The Free plan allows 3 requests per second per account.
	// https://circleback.ai/docs/api#rate-limits
	maxConcurrentTranscriptFetches = 3

	// A rate-limited transcript request is retried after its Retry-After wait, up to this many tries.
	maxTranscriptAttempts = 3
)

// withTranscripts adds each meeting's transcript to its fields: the list of
// {speaker, text, timestamp (seconds from the start)} segments, in chronological order.
// The meetings list has no transcript, so it is fetched per meeting, for the meetings
// that passed the time filter only. Raw stays the list record.
// https://circleback.ai/docs/api (GET /meeting/{meetingId}/transcript)
func (c *Connector) withTranscripts(
	ctx context.Context, marshal common.MarshalFromNodeFunc,
) common.MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		rows, err := marshal(records, fields)
		if err != nil {
			return nil, err
		}

		transcripts, err := c.fetchTranscripts(ctx, rows)
		if err != nil {
			return nil, err
		}

		for index := range rows {
			rows[index].Fields[fieldTranscript] = transcripts[rows[index].Id]
		}

		return rows, nil
	}
}

func (c *Connector) fetchTranscripts(
	ctx context.Context, rows []common.ReadResultRow,
) (map[string]any, error) {
	var mutex sync.Mutex

	transcripts := make(map[string]any, len(rows))
	jobs := make([]simultaneously.Job, 0, len(rows))

	for _, row := range rows {
		meetingID := row.Id

		jobs = append(jobs, func(ctx context.Context) error {
			segments, err := c.fetchTranscript(ctx, meetingID)
			if err != nil {
				return err
			}

			mutex.Lock()
			defer mutex.Unlock()

			transcripts[meetingID] = segments

			return nil
		})
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentTranscriptFetches, jobs...); err != nil {
		return nil, err
	}

	return transcripts, nil
}

// fetchTranscript returns a meeting's transcript segments, or nil when Circleback has none for it.
//
// A meeting with no transcript (no recording, or still processing) must not fail the page, so a
// 400 or 404 for one meeting reads as no transcript. A 429 is waited out for the seconds in its
// Retry-After header and retried: on the Free plan (20 requests per minute) a page of 20 meetings
// with their transcripts needs more requests than one minute allows.
// https://circleback.ai/docs/api#rate-limits
func (c *Connector) fetchTranscript(ctx context.Context, meetingID string) (any, error) {
	if meetingID == "" {
		return nil, fmt.Errorf("%w: meeting id", common.ErrMissingExpectedValues)
	}

	transcriptURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, "meeting", meetingID, "transcript")
	if err != nil {
		return nil, err
	}

	for attempt := 1; ; attempt++ {
		response, err := c.JSONHTTPClient().Get(ctx, transcriptURL.String())
		if err == nil {
			segments, err := common.UnmarshalJSON[[]any](response)
			if err != nil {
				return nil, err
			}

			return *segments, nil
		}

		if errors.Is(err, common.ErrBadRequest) {
			return nil, nil
		}

		var limited *rateLimitError
		if !errors.As(err, &limited) || attempt == maxTranscriptAttempts {
			return nil, err
		}

		if err := sleep(ctx, limited.retryAfter); err != nil {
			return nil, err
		}
	}
}

func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
