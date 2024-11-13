package klaviyo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/klaviyo/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var prioritySinceFieldsForRead = []string{ //nolint:gochecknoglobals
	"updated_at", // most desired field for incremental readin
	"updated",
	"datetime",
	"completed_at", // for Bulk Jobs
	"created_at",
	"created", // least preferred field
}

var objectsNameToSinceFieldName = make(map[common.ModuleID]map[string]string) //nolint:gochecknoglobals

func init() {
	// Every object should be associated with a fieldName which will be used for iterative reading via Since parameter.
	// Instead of recalculating this over and over on package start we can "find" fields of interest.
	// There is a preferred choice when it comes to time filtering represented by `prioritySinceFieldsForRead`.
	for moduleID, module := range metadata.Schemas.Modules {
		objectsNameToSinceFieldName[moduleID] = make(map[string]string)

		for objectName, object := range module.Objects {
		Search:
			for _, preferredSinceField := range prioritySinceFieldsForRead {
				for currentField := range object.FieldsMap {
					if preferredSinceField == currentField {
						objectsNameToSinceFieldName[moduleID][objectName] = preferredSinceField

						// break search for this object
						break Search
					}
				}
			}
		}
	}
}
