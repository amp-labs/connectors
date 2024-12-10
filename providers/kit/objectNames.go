package kit

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameBroadCasts     = "broadcasts"
	objectNameCustomFields   = "custom_fields"
	objectNameEmailTemplates = "email_templates"
	objectNameForms          = "forms"
	objectNamePurchases      = "purchases"
	objectNameSequences      = "sequences"
	objectNameSegments       = "segments"
	objectNameSubscribers    = "subscribers"
	objectNameTags           = "tags"
	objectNameWebhooks       = "webhooks"
)

var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNameBroadCasts,
	objectNameCustomFields,
	objectNameEmailTemplates,
	objectNameTags,
	objectNameForms,
	objectNamePurchases,
	objectNameSequences,
	objectNameSegments,
	objectNameSubscribers,
	objectNameWebhooks,
)
