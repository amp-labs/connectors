package kit

import "github.com/amp-labs/connectors/common/handy"

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

var supportedObjectsByRead = handy.NewSet( //nolint:gochecknoglobals
	objectNameCustomfields,
	objectNameEmailtemplates,
	objectNameTags,
)
