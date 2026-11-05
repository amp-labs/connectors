package associations

import (
	"context"
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

type (
	recordItem                  map[string]any
	recordId                    = string
	recordItems                 map[recordId]recordItem
	meetingAssociationCollector func(ctx context.Context, base *core.Base, meetingId string) (recordItems, error)
)

// meetingAssociationsToObject fetches an association for multiple Meetings
// concurrently and groups the resulting records by their parent Meeting ID.
//
// Each returned association contains the associated record's identifier and
// raw response data. If fetching any association fails, the collected errors
// are returned together.
func meetingAssociationsToObject(ctx context.Context,
	base *core.Base,
	meetingIds []readhelper.RowID,
	fetchRecords meetingAssociationCollector,
) (map[readhelper.RowID][]common.Association, error) {
	tasks := make([]parallelfetch.Task[string, recordItems], len(meetingIds))
	for index, meetingId := range meetingIds {
		tasks[index] = func(ctx context.Context) (taskID string, data *recordItems, err error) {
			records, err := fetchRecords(ctx, base, meetingId)
			if err != nil {
				return "", nil, err
			}

			return meetingId, &records, nil
		}
	}

	result := parallelfetch.Execute(ctx, tasks, core.MaxReadConcurrency)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	registry := make(map[readhelper.RowID][]common.Association)
	for meetingId, records := range result.Records {
		registry[meetingId] = make([]common.Association, len(records))

		idx := 0
		for objectId, record := range records {
			registry[meetingId][idx] = common.Association{
				ObjectId:                    objectId,
				Raw:                         record,
				ProviderAssociationMetadata: nil,
			}
			idx += 1
		}
	}

	return registry, nil
}

// https://developers.zoom.us/docs/api/meetings/#tag/meetings/get/past_meetings/{meetingId}/participants
func fetchMeetingParticipants(ctx context.Context,
	base *core.Base,
	meetingId string,
) (recordItems, error) {
	url, err := base.GetMeetingParticipantUrl(meetingId)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("page_size", core.DefaultPageSize)

	return fetchItems(ctx, base, url, "participants", "id")
}

// Transcript is a single object rather than a collection, so it is treated
// as additional metadata for a Meeting.
// https://developers.zoom.us/docs/api/meetings/#tag/meeting-transcript/get/meetings/{meetingId}/transcript
func fetchMeetingTranscript(ctx context.Context,
	base *core.Base,
	meetingId string,
) (recordItems, error) {
	url, err := base.GetMeetingTranscriptUrl(meetingId)
	if err != nil {
		return nil, err
	}

	item, err := fetchItem(ctx, base, url, "meeting_id")
	if err != nil {
		if errors.Is(err, common.ErrNotFound) && strings.Contains(err.Error(), "error code (3322)") {
			// This meeting transcript does not exist.
			// No associations to return.
			return make(recordItems), nil
		}

		return nil, err
	}

	return item, nil
}

// Summary is an additional metadata for a Meeting.
// https://developers.zoom.us/docs/api/meetings/#tag/summaries/get/meetings/{meetingId}/meeting_summary
func fetchMeetingSummary(ctx context.Context, base *core.Base, meetingId string) (recordItems, error) {
	url, err := base.GetMeetingSummaryUrl(meetingId)
	if err != nil {
		return nil, err
	}

	return fetchItem(ctx, base, url, "meeting_id")
}
