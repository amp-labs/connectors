package associations

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

const (
	MeetingParticipants = "participants"
	MeetingTranscripts  = "transcript"
	MeetingSummaries    = "summary"
)

// CreateDataMarshallerWithAssociations creates a marshaller that hydrates with Meetings associations.
func CreateDataMarshallerWithAssociations(
	ctx context.Context, base *core.Base,
	objectName common.ObjectName, associatedObjects []string,
	identifier string,
) common.MarshalFromNodeFunc {
	return readhelper.ChainedMarshaller(
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField(identifier)),
		readhelper.HydrateAssociations(ctx, objectName, associatedObjects, lookupAssociations(base)),
	)
}

// Currently, only Meetings are supported.
func lookupAssociations(base *core.Base) readhelper.AssociationLookup {
	return func(ctx context.Context,
		object common.ObjectName,
		identifiers []readhelper.RowID,
		toObject string,
	) (map[readhelper.RowID][]common.Association, error) {
		if core.ObjectNameMeetings.Equals(object.String()) {
			return meetingAssociations(ctx, toObject, base, identifiers)
		}

		return nil, readhelper.ErrAssociationLookupNotImplemented
	}
}

func meetingAssociations(ctx context.Context,
	toObject string,
	base *core.Base,
	ids []readhelper.RowID,
) (map[readhelper.RowID][]common.Association, error) {
	switch toObject {
	case MeetingParticipants:
		return meetingAssociationsToObject(ctx, base, ids, fetchMeetingParticipants)
	case MeetingTranscripts:
		return meetingAssociationsToObject(ctx, base, ids, fetchMeetingTranscript)
	case MeetingSummaries:
		return meetingAssociationsToObject(ctx, base, ids, fetchMeetingSummary)
	default:
		return nil, readhelper.ErrAssociationsUnsupported
	}
}
