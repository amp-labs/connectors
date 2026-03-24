package revenuecat

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// writeAndDeleteSupportedObjects are object names that support write and delete.
// Subscriptions and other read-only objects are excluded. Validated in buildWriteRequest and buildDeleteRequest.
//
//nolint:gochecknoglobals
var writeAndDeleteSupportedObjects = datautils.NewStringSet(
	"apps",
	"customers",
	"entitlements",
	"integrations_webhooks",
	"offerings",
	"products",
)
