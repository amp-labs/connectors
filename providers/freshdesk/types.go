package freshdesk

import "slices"

// readSupportedObjects represents a list of objects supported by Read Connector.
var readSupportedObjects = []string{ //nolint:gochecknoglobals
	"contacts",
	"tickets",
	"ticket-forms",
	"agents",
	"roles",
	"groups",
	"companies",
	"canned_response_folders",
	"surveys",
	"time_entries",
	"email_configs",
	"products",
	"business_hours",
	"scenario_automations",
	"sla_policies",
}

// objectReadPath represents a mapping of an object to it's read path.
var objectReadPath = map[string]string{ //nolint:gochecknoglobals
	"mailboxes": "email/mailboxes",
	"settings":  "settings/helpdesk",
	"skills":    "admin/skills",
}

func objectReadSupported(objectName string) bool {
	if _, exists := objectReadPath[objectName]; exists {
		return exists
	}

	return slices.Contains(readSupportedObjects, objectName)
}
