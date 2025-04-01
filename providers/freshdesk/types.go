package freshdesk

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// readSupportedObjects represents a list of objects supported by Read Connector.
var readSupportedObjects = datautils.NewSet( //nolint:gochecknoglobals
	"agents",
	"business_hours",
	"canned_response_folders",
	"companies",
	"company-fields",
	"contacts",
	"contact-fields",
	"email_configs",
	"groups",
	"mailboxes",
	"products",
	"roles",
	"scenario_automations",
	"settings",
	"skills",
	"sla_policies",
	"surveys",
	"ticket-fields",
	"ticket-forms",
	"tickets",
	"time_entries",
)

var writeSupportedObjects = datautils.NewSet( //nolint:gochecknoglobals
	"agents",
	"canned_response_folders",
	"companies",
	"company-fields",
	"contact-activities",
	"contact-fields",
	"contacts",
	"groups",
	"mailboxes",
	"skills",
	"sla_policies",
	"thread",
	"ticket-fields",
	"ticket-forms",
	"tickets",
)

// objectResourcePath represents a mapping of an object to it's read/write resource.
var objectResourcePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"mailboxes":     "email/mailboxes",
	"settings":      "settings/helpdesk",
	"skills":        "admin/skills",
	"thread":        "collaboration/threads",
	"ticket-fields": "admin/ticket_fields",
}, func(objectName string) (path string) {
	return objectName
})
