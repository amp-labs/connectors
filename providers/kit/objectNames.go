package kit

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameBroadcasts     = "broadcasts"
	objectNameCustomfields   = "custom_fields"
	objectNameEmailtemplates = "email_templates"
	objectNameForms          = "forms"
	objectNamePurchases      = "purchases"
	objectNameSequences      = "sequences"
	objectNameSegments       = "segments"
	objectNameSubscribers    = "subscribers"
	objectNameTags           = "tags"
	objectNameWebhooks       = "webhooks"
)

var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCustomfields,
	objectNameEmailtemplates,
	objectNameTags,
)
