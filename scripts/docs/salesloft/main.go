package main

import (
	"fmt"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/salesloft/metadata"
)

// Prints supported read objects in random order formatted for https://github.com/amp-labs/docs.
// nolint:forbidigo
func main() {
	for objectName, object := range metadata.Schemas.Modules[staticschema.RootModuleID].Objects {
		fmt.Printf("- [%v](%v) (%v)\n", object.DisplayName, *object.DocsURL, objectName)
	}
}
