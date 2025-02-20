package google

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/google/metadata"
)

const (
	objectNameCalendarList  = "calendarList"
	objectNameMyConnections = "myConnections"
	objectNameContactGroups = "contactGroups"
	objectNameOtherContacts = "otherContacts" // read only - TODO need to test READ
	// This object seems to be for Directory Admins.
	// https://support.google.com/a/answer/1628009?hl=en
	objectNamePeopleDirectory = "peopleDirectory" // read only
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByCreate = common.ModuleObjectNameToURLPath{ //nolint:gochecknoglobals
	ModuleCalendar: datautils.NewDefaultMap(map[string]string{
		objectNameCalendarList: "/calendar/v3/users/me/calendarList",
	}, emptyURL),
	ModulePeople: datautils.NewDefaultMap(map[string]string{
		// https://developers.google.com/people/api/rest/v1/people/createContact
		objectNameMyConnections: "/v1/people:createContact", // TODO testing
		objectNameContactGroups: "/v1/contactGroups",
	}, emptyURL),
}

var supportedObjectsByUpdate = common.ModuleObjectNameToURLPath{ //nolint:gochecknoglobals
	ModuleCalendar: datautils.NewDefaultMap(map[string]string{
		objectNameCalendarList: "/calendar/v3/users/me/calendarList",
	}, emptyURL),
	ModulePeople: datautils.NewDefaultMap(map[string]string{
		// https://developers.google.com/people/api/rest/v1/people/updateContact
		objectNameMyConnections: "/v1/people:updateContact", // TODO testing MUST BE PATCH!!!!!!!!!!!!!!!!!!!
		objectNameContactGroups: "/v1/contactGroups",
	}, emptyURL),
}

var supportedObjectsByDelete = common.ModuleObjectNameToURLPath{ //nolint:gochecknoglobals
	ModuleCalendar: datautils.NewDefaultMap(map[string]string{
		objectNameCalendarList: "/calendar/v3/users/me/calendarList",
	}, emptyURL),
	ModulePeople: datautils.NewDefaultMap(map[string]string{
		// https://developers.google.com/people/api/rest/v1/people/deleteContact
		objectNameMyConnections: "/v1/people:deleteContact", // TODO testing
		objectNameContactGroups: "/v1/contactGroups",
	}, emptyURL),
}

// resourceIdentifierFormat breaks resourceName into parts, where format is: "name/identifier".
func resourceIdentifierFormat(resourceName string) (objectName string, recordID string, ok bool) {
	parts := strings.Split(resourceName, "/")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func emptyURL(string) (urlPath string) {
	return ""
}
