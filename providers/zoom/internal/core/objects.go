package core

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/zoom/internal/metadata"
)

// FieldUserId is the field name used to store the ID of the user from whom
// a record was retrieved. It is injected into ReadResult.Data[*].Fields for
// records belonging to ObjectCategoryAllUsers to preserve their source
// user context in the aggregated result.
const FieldUserId = "AMPERSAND-userId"

var (
	ObjectNameContactGroup  = "contacts_groups" // nolint: gochecknoglobals
	ObjectNameUser          = "users"           // nolint: gochecknoglobals
	ObjectNameGroup         = "groups"          // nolint: gochecknoglobals
	ObjectNameTrackingField = "tracking_fields" // nolint: gochecknoglobals
	ObjectNameDevice        = "devices"         // nolint: gochecknoglobals
	ObjectNameH322Device    = "h323_devices"    // nolint: gochecknoglobals
)

// UserScopedObject holds the object names for the two supported retrieval
// modes of a user-scoped Zoom resource:
//   - the current user (Self)
//   - and all users in the account (All).
type UserScopedObject struct {
	Self string
	All  string
}

func (o UserScopedObject) Equals(objectName string) bool {
	return o.All == objectName || o.Self == objectName
}

// nolint:gochecknoglobals
var (
	ObjectNameMeetingTemplates = UserScopedObject{Self: "meeting_templates", All: "account_meeting_templates"}
	ObjectNameMeetings         = UserScopedObject{Self: "meetings", All: "account_meetings"}
	ObjectNameMeetingsReport   = UserScopedObject{Self: "meetings_report", All: "account_meetings_report"}
	ObjectNameRecordings       = UserScopedObject{Self: "recordings", All: "account_recordings"}
	ObjectNameTsp              = UserScopedObject{Self: "tsp", All: "account_tsp"}
	ObjectNameUpcomingMeetings = UserScopedObject{Self: "upcoming_meetings", All: "account_upcoming_meetings"}
	ObjectNameWebinars         = UserScopedObject{Self: "webinars", All: "account_webinars"}

	// ObjectCategoryAllUsers contains connector-defined object categories
	// that aggregate records across all users in a Zoom account.
	// These objects do not correspond to distinct endpoints in the Zoom API.
	// Instead, they represent connector-level abstractions that combine records retrieved for individual users.
	ObjectCategoryAllUsers = datautils.NewSetFromList([]string{
		ObjectNameMeetingTemplates.All,
		ObjectNameMeetings.All,
		ObjectNameMeetingsReport.All,
		ObjectNameRecordings.All,
		ObjectNameTsp.All,
		ObjectNameUpcomingMeetings.All,
		ObjectNameWebinars.All,
	})
)

var (
	SupportedObjectsByRead  = metadata.Catalog.Objects.KeySet()  // nolint: gochecknoglobals
	SupportedObjectsByWrite = objectNameToWritePath.Map.KeySet() // nolint: gochecknoglobals
)
