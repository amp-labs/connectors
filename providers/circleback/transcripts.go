package circleback

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/amp-labs/amp-common/simultaneously"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

const (
	// fieldTranscript is the meetings field holding the transcript, which the list does not return.
	fieldTranscript = "transcript"

	// At most this many transcript requests are in flight: the Free plan's per-second limit.
	// https://circleback.ai/docs/api#rate-limits
	maxConcurrentTranscriptFetches = 3

	// A rate-limited transcript request is tried this many times in all.
	maxTranscriptAttempts = 3
)

var errNoTranscript = errors.New("meeting has no transcript")

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
			var transcript any

			segments, err := c.fetchTranscript(ctx, meetingID)

			switch {
			case err == nil:
				transcript = segments
			case !errors.Is(err, errNoTranscript):
				return err
			}

			mutex.Lock()
			defer mutex.Unlock()

			transcripts[meetingID] = transcript

			return nil
		})
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentTranscriptFetches, jobs...); err != nil {
		return nil, err
	}

	return transcripts, nil
}

// fetchTranscript returns a meeting's transcript segments, or errNoTranscript when Circleback has none.
//
// A meeting with no transcript (no recording, or still processing) must not fail the page, so a
// 400 or 404 for one meeting reads as no transcript; every other failure is an error. Requests wait
// while the account's rate-limit window is used up, and a 429 is retried once its Retry-After
// has passed (see ratelimit.go).
// https://circleback.ai/docs/api#rate-limits
func (c *Connector) fetchTranscript(ctx context.Context, meetingID string) ([]any, error) {
	if meetingID == "" {
		return nil, fmt.Errorf("%w: meeting id", common.ErrMissingExpectedValues)
	}

	transcriptURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, "meeting", meetingID, "transcript")
	if err != nil {
		return nil, err
	}

	for attempt := 1; ; attempt++ {
		if err := c.pacer.wait(ctx); err != nil {
			return nil, err
		}

		response, err := c.JSONHTTPClient().Get(ctx, transcriptURL.String())
		if err == nil {
			c.pacer.observe(response.Headers)

			segments, err := common.UnmarshalJSON[[]any](response)
			if err != nil {
				return nil, err
			}

			return *segments, nil
		}

		if err = transcriptFailure(err, attempt); err != nil {
			return nil, err
		}
	}
}

// transcriptFailure decides what a failed transcript request means: errNoTranscript for a 400 or
// 404, nil when a 429 should be tried again, and the error itself otherwise.
func transcriptFailure(err error, attempt int) error {
	var httpErr *common.HTTPError
	if !errors.As(err, &httpErr) {
		return err
	}

	switch {
	case httpErr.Status == http.StatusBadRequest || httpErr.Status == http.StatusNotFound:
		return errNoTranscript
	case httpErr.Status == http.StatusTooManyRequests && attempt < maxTranscriptAttempts:
		return nil
	default:
		return err
	}
}
