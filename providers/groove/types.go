package groove

import "github.com/amp-labs/connectors/internal/datautils"

var readSupportedObjects = datautils.NewSet( //nolint:gochecknoglobals
	"tickets",
	"customers",
	"tickets/count",
	"mailboxes",
	"folders",
	"agents",
	"groups",
	"kb", // represents knowledge base.
	"kb/themes",
	"widgets",
)

var responseFieldMap = map[string]string{ //nolint:gochecknoglobals
	"tickets":       "tickets",
	"customers":     "customers",
	"tickets/count": "",
	"mailboxes":     "mailboxes",
	"folders":       "folders",
	"agents":        "agents",
	"groups":        "groups",
	"kb":            "knowledge_bases",
	"kb/themes":     "themes",
	"widgets":       "widgets",
}
