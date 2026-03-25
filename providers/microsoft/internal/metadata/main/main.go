package main

import (
	"fmt"
	"sort"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/microsoft/internal/metadata"
)

var priority = []string{ // nolint:gochecknoglobals
	"lastModifiedDateTime", // highest
	"lastEditedDateTime",
	"lastUpdatedDateTime",
	"lastUpdateDateTime",
	"modifiedDateTime",
	"creationDateTime",
	"createdDateTime", // lowest
}

// Create mapping between object name and the field property that is most suitable for incremental reading.
func main() {
	result := datautils.Map[string, string]{}

	for objectName, object := range metadata.Schemas.Modules[common.ModuleRoot].Objects {
		fields := datautils.Map[string, staticschema.FieldMetadata](object.Fields)
		for _, fieldName := range priority {
			if fields.Has(fieldName) {
				result[objectName] = fieldName

				break
			}
		}
	}

	keys := result.Keys()
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("\"%v\": \"%v\",\n", key, result[key]) // nolint:forbidigo
	}

	// As of now 268/568 ~~ 47%.
	fmt.Printf("\nNumber objects that can use incremental reading: %v\n", len(result)) // nolint:forbidigo
}
