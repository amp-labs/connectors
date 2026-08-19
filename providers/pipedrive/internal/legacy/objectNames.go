package legacy

import "github.com/amp-labs/connectors/internal/datautils"

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{}, //nolint:gochecknoglobals
	func(key string) string {
		return key
	},
)

var notesFlagFields = datautils.NewSet("pinned_to_deal_flag", "pinned_to_person_flag", // nolint: gochecknoglobals
	"pinned_to_organization_flag", "pinned_to_lead_flag")

var metadataDiscoveryEndpoints = datautils.Map[string, string]{ // nolint: gochecknoglobals
	"activities":    "activityFields",
	"deals":         "dealFields",
	"products":      "productFields",
	"persons":       "personFields",
	"organizations": "organizationFields",
	"notes":         "noteFields",
}

// objectNameAliases maps alternative object names onto the name they are stored under in
// the static schema file. Reads and metadata resolve the schema and URL path through this,
// so callers may use either spelling. The name the caller supplied is never rewritten:
// it is what comes back in the metadata result and it is what they keep referring to.
var objectNameAliases = datautils.Map[string, string]{ // nolint: gochecknoglobals
	"mailThreads": objectNameMailThreads,
}

// canonicalObjectName resolves an object name to the name used in the static schema file.
// Names that are not aliases are returned unchanged.
func canonicalObjectName(objectName string) string {
	if canonical, ok := objectNameAliases[objectName]; ok {
		return canonical
	}

	return objectName
}

const (
	objectNameMailThreads = "mailboxes/mailThreads"

	// mailboxFolderInbox is the folder read for mail threads. The API accepts inbox,
	// drafts, sent or archive and already defaults to inbox; we only read the inbox.
	// https://developers.pipedrive.com/docs/api/v1/Mailbox
	mailboxFolderInbox = "inbox"
)
