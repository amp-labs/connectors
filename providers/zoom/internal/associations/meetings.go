package associations

import (
	"context"
	"errors"

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

// https://developers.zoom.us/docs/api/meetings/#tag/meetings/get/past_meetings/%7BmeetingId%7D/participants
func fetchMeetingParticipants(ctx context.Context,
	base *core.Base,
	meetingId string,
) (recordItems, error) {
	url, err := base.GetMeetingParticipantUrl(meetingId)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("page_size", core.DefaultPageSize)

	return fetchPagedParticipants(ctx, base, url)
}

// Meeting transcript is a transcript for the latest meeting instance.
// https://developers.zoom.us/docs/api/meetings/#tag/meeting-transcript/get/meetings/%7BmeetingId%7D/transcript
func fetchMeetingTranscript(latestMeetingInstances map[string]string) meetingAssociationCollector {
	return func(ctx context.Context,
		base *core.Base,
		meetingId string,
	) (recordItems, error) {
		meetingInstanceId, ok := latestMeetingInstances[meetingId]
		if !ok {
			return make(recordItems), nil
		}

		url, err := base.GetMeetingTranscriptUrl(meetingInstanceId)
		if err != nil {
			return nil, err
		}

		// meeting_id represents a UUID of a meeting instance.
		return fetchItem(ctx, base, url, "meeting_id")
	}
}

// Meeting summary is a summary for the latest meeting instance.
// https://developers.zoom.us/docs/api/meetings/#tag/summaries/get/meetings/%7BmeetingId%7D/meeting_summary
func fetchMeetingSummary(latestMeetingInstances map[string]string) meetingAssociationCollector {
	return func(ctx context.Context, base *core.Base, meetingId string) (recordItems, error) {
		meetingInstanceId, ok := latestMeetingInstances[meetingId]
		if !ok {
			return make(recordItems), nil
		}

		url, err := base.GetMeetingSummaryUrl(meetingInstanceId)
		if err != nil {
			return nil, err
		}

		// meeting_uuid represents a UUID of a meeting instance.
		return fetchItem(ctx, base, url, "meeting_uuid")
	}
}
