package associations

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/pagedbatch"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

type (
	recordItem  map[string]any
	recordId    = string
	recordItems map[recordId]recordItem
)

// meetingAssociation is how one kind of meeting association is fetched: the URL of its first
// request for a meeting, and how its responses are read.
type meetingAssociation struct {
	url  func(base *core.Base, meetingId string) (*urlbuilder.URL, error)
	read readFunc
}

// meetingAssociationsToObject fetches an association for multiple Meetings as request batches, and
// groups the resulting records by their parent Meeting ID. Paginated associations go out page by
// page: the first page for every meeting, then the next page for every meeting that has one.
//
// Each returned association contains the associated record's identifier and
// raw response data. If fetching any association fails, that error is returned.
func meetingAssociationsToObject(ctx context.Context,
	base *core.Base,
	meetingIds []readhelper.RowID,
	association meetingAssociation,
) (map[readhelper.RowID][]common.Association, error) {
	first := make([]common.BatchRequest, len(meetingIds))
	records := make([]recordItems, len(meetingIds))

	for index, meetingId := range meetingIds {
		url, err := association.url(base, meetingId)
		if err != nil {
			return nil, err
		}

		first[index] = pagedbatch.Get(url.String())
		records[index] = make(recordItems)
	}

	err := pagedbatch.Run(ctx, base.JSONHTTPClient(), first,
		func(walk int, rsp common.BatchResponse) (common.BatchRequest, bool, error) {
			next, err := association.read(rsp, records[walk])

			switch {
			case err != nil:
				return pagedbatch.Fail(err)
			case next == "":
				return pagedbatch.Done()
			default:
				return pagedbatch.Next(next)
			}
		})
	if err != nil {
		return nil, err
	}

	registry := make(map[readhelper.RowID][]common.Association)

	for index, meetingId := range meetingIds {
		associations := make([]common.Association, 0, len(records[index]))

		for objectId, record := range records[index] {
			associations = append(associations, common.Association{
				ObjectId:                    objectId,
				Raw:                         record,
				ProviderAssociationMetadata: nil,
			})
		}

		registry[meetingId] = associations
	}

	return registry, nil
}

// https://developers.zoom.us/docs/api/meetings/#tag/meetings/get/past_meetings/{meetingId}/participants
//
//nolint:gochecknoglobals
var meetingParticipants = meetingAssociation{
	url: func(base *core.Base, meetingId string) (*urlbuilder.URL, error) {
		url, err := base.GetMeetingParticipantUrl(meetingId)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam("page_size", core.DefaultPageSize)

		return url, nil
	},
	read: readItems("participants", "id"),
}

// Transcript is a single object rather than a collection, so it is treated
// as additional metadata for a Meeting.
// https://developers.zoom.us/docs/api/meetings/#tag/meeting-transcript/get/meetings/{meetingId}/transcript
//
//nolint:gochecknoglobals
var meetingTranscript = meetingAssociation{
	url: func(base *core.Base, meetingId string) (*urlbuilder.URL, error) {
		return base.GetMeetingTranscriptUrl(meetingId)
	},
	read: readTranscript,
}

// Summary is an additional metadata for a Meeting.
// https://developers.zoom.us/docs/api/meetings/#tag/summaries/get/meetings/{meetingId}/meeting_summary
//
//nolint:gochecknoglobals
var meetingSummary = meetingAssociation{
	url: func(base *core.Base, meetingId string) (*urlbuilder.URL, error) {
		return base.GetMeetingSummaryUrl(meetingId)
	},
	read: readItem("meeting_id"),
}
