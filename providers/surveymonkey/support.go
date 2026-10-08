package surveymonkey

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

const apiVersion = "v3"

// writeAndDeleteSupportedObjects support create, update, and delete.
//
//nolint:gochecknoglobals
var writeAndDeleteSupportedObjects = datautils.NewStringSet(
	objectContacts,
	objectContactLists,
	objectSurveys,
)

// writeCreateOnlyObjects support create only.
//
//nolint:gochecknoglobals
var writeCreateOnlyObjects = datautils.NewStringSet(
	objectSurveyFolders,
)

// writeUpdateOnlyObjects support update only.
//
//nolint:gochecknoglobals
var writeUpdateOnlyObjects = datautils.NewStringSet(
	objectContactFields,
)

// writeSupportedObjects is the union of all write-capable objects.
//
//nolint:gochecknoglobals
var writeSupportedObjects = datautils.MergeSets(
	writeAndDeleteSupportedObjects,
	writeCreateOnlyObjects,
	writeUpdateOnlyObjects,
)

func validateWriteOperation(params common.WriteParams) error {
	if !writeSupportedObjects.Has(params.ObjectName) {
		return common.ErrOperationNotSupportedForObject
	}

	if params.IsCreate() && writeUpdateOnlyObjects.Has(params.ObjectName) {
		return common.ErrOperationNotSupportedForObject
	}

	if params.IsUpdate() && writeCreateOnlyObjects.Has(params.ObjectName) {
		return common.ErrOperationNotSupportedForObject
	}

	return nil
}
