package wealthbox

// documentTypeByObjectName maps a connector object name to the `document_type` value
// used by the custom fields endpoint (GET /v1/categories/custom_fields?document_type=X).
// https://dev.wealthbox.com/#topics-custom-fields
//
// Notes are intentionally absent: Wealthbox does not support custom fields on notes.
var documentTypeByObjectName = map[string]string{ // nolint:gochecknoglobals
	objectNameContacts:      "Contact",
	objectNameTasks:         "Task",
	objectNameEvents:        "Event",
	objectNameOpportunities: "Opportunity",
	objectNameProjects:      "Project",
}

const (
	objectNameContacts      = "contacts"
	objectNameTasks         = "tasks"
	objectNameEvents        = "events"
	objectNameOpportunities = "opportunities"
	objectNameProjects      = "projects"
	objectNameNotes         = "notes"
)
