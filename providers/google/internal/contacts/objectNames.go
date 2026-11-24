package contacts

import (
	"net/http"

	"github.com/amp-labs/connectors/providers/google/internal/shared"
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
var endpoints = shared.Endpoints{ // nolint:gochecknoglobals
	shared.OperationCreate: {
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
	shared.OperationUpdate: {
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
	shared.OperationDelete: {
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
