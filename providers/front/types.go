package front

import "slices"

// supportedReadObjects represents objects supported by Front Read Connector.
var supportedReadObjects = []string{ //nolint:gochecknoglobals
	"accounts",
	"channels",
	"contacts",
	"contact_lists",
	"conversations",
	"events",
	"inboxes",
	"knowledge_bases",
	"links",
	"rules",
	"message_templates",
	"message_template_folders",
	"shifts",
	"tags",
	"teammates",
	"teammate_groups",
	"teams",
}

func supportsRead(objectName string) bool {
	return slices.Contains(supportedReadObjects, objectName)
}
