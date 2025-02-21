package google

import (
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
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

var supportedObjectsByCreate = common.NewModuleObjectNameToOperationDescription(http.MethodPost,
	map[common.ModuleID]map[string]common.OperationDescription{
		ModuleCalendar: {
			objectNameCalendarList: {
				Path: "/calendar/v3/users/me/calendarList",
			},
		},
		ModulePeople: {
			objectNameMyConnections: {
				// https://developers.google.com/people/api/rest/v1/people/createContact
				Path: "/v1/people:createContact", // TODO testing
			},
			objectNameContactGroups: {
				Path: "/v1/contactGroups",
			},
		},
	},
)

var supportedObjectsByUpdate = common.NewModuleObjectNameToOperationDescription(http.MethodPut,
	map[common.ModuleID]map[string]common.OperationDescription{
		ModuleCalendar: {
			objectNameCalendarList: {
				Path: "/calendar/v3/users/me/calendarList",
			},
		},
		ModulePeople: {
			objectNameMyConnections: {
				// https://developers.google.com/people/api/rest/v1/people/updateContact
				Operation: http.MethodPatch,
				Path:      "/v1/people/{{.recordID}}:updateContact", // TODO testing MUST BE PATCH!!!!!!!!!!!!!!!!!!!
			},
			objectNameContactGroups: {
				Path: "/v1/contactGroups",
			},
		},
	},
)

var supportedObjectsByDelete = common.NewModuleObjectNameToOperationDescription(http.MethodDelete,
	map[common.ModuleID]map[string]common.OperationDescription{
		ModuleCalendar: {
			objectNameCalendarList: {
				Path: "/calendar/v3/users/me/calendarList",
			},
		},
		ModulePeople: {
			objectNameMyConnections: {
				// https://developers.google.com/people/api/rest/v1/people/deleteContact
				Operation: http.MethodPatch,
				Path:      "/v1/people/{{.recordID}}:deleteContact", // TODO testing
			},
			objectNameContactGroups: {
				Path: "/v1/contactGroups",
			},
		},
	},
)

// resourceIdentifierFormat breaks resourceName into parts, where format is: "name/identifier".
func resourceIdentifierFormat(resourceName string) (objectName string, recordID string, ok bool) {
	parts := strings.Split(resourceName, "/")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}
