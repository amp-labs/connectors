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

// supportedWriteObjects represents objects supported by Front Write Connector.
var supportedCreationObjects = []string{ //nolint:gochecknoglobals
	"accounts",
	"contacts",
	"contact_lists",
	"conversations",
	"inboxes", // Create an inbox in the default team (workspace)
	"knowledge_bases",
	"links",
	"message_templates",
	"message_template_folders",
	"shifts",
	"tags",
	"teammate_groups",
	"teams",
}

var supportedPatchObjects = []string{ //nolint:gochecknoglobals
	"accounts",
	"channels",
	"comments",
	"contacts",
	"conversations",
	"links",
	"message_templates",
	"message_template_folders",
	"shifts",
	"signatures",
	"tags",
	"teammates",
	"teammate_groups",
}

func supportsRead(objectName string) bool {
	return slices.Contains(supportedReadObjects, objectName)
}

func supportsCreation(objectName string) bool {
	return slices.Contains(supportedCreationObjects, objectName)
}

func supportsPatching(objectName string) bool {
	return slices.Contains(supportedPatchObjects, objectName)
}
