package contacts

import (
	"net/http"

	"github.com/amp-labs/connectors/providers/google/internal/core"
)

const (
	objectNameMyConnections = "myConnections"
	objectNameContactGroups = "contactGroups"
	objectNameOtherContacts = "otherContacts"
	// This object seems to be for Directory Admins.
	// https://support.google.com/a/answer/1628009?hl=en
	objectNamePeopleDirectory = "peopleDirectory" // read only
)

// Maps object names to URL endpoints.
var endpoints = core.Endpoints{ // nolint:gochecknoglobals
	core.OperationCreate: {
		objectNameContactGroups: {
			Method: http.MethodPost,
			Path:   "/contactGroups",
		},
		objectNameMyConnections: {
			Method: http.MethodPost,
			// https://developers.google.com/people/api/rest/v1/people/createContact
			Path: "/people:createContact",
		},
	},
	core.OperationUpdate: {
		objectNameContactGroups: {
			Method: http.MethodPut,
			Path:   "/contactGroups",
		},
		objectNameMyConnections: {
			// https://developers.google.com/people/api/rest/v1/people/updateContact
			Method: http.MethodPatch,
			Path:   "/people/{{.recordID}}:updateContact",
		},
	},
	core.OperationDelete: {
		objectNameContactGroups: {
			Path: "/contactGroups",
		},
		objectNameMyConnections: {
			// https://developers.google.com/people/api/rest/v1/people/deleteContact
			Method: http.MethodPatch,
			Path:   "/people/{{.recordID}}:deleteContact",
		},
	},
}
