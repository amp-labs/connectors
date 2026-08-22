package surveymonkey

import "github.com/amp-labs/connectors/internal/datautils"

const apiVersion = "v3"

// writeAndDeleteSupportedObjects are object names that support write and delete.
//
//nolint:gochecknoglobals
var writeAndDeleteSupportedObjects = datautils.NewStringSet(
	objectContacts,
	objectContactLists,
)
