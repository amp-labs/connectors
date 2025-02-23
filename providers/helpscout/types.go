package helpscout

import "github.com/amp-labs/connectors/internal/datautils"

var supportedReadObjects = datautils.NewSet( //nolint:gochecknoglobals
	"conversations",
	"customers",
	"mailboxes",
	"customer-properties",
	"tags",
	"teams",
	"users",
	"webhooks",
	"workflows",
)
