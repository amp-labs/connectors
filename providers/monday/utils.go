package monday

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/spyzhov/ajson"
)

const (
	objectBoard = "boards"
	objectItem  = "items"
	objectUser  = "users"
)

// Map of object names to their ID field paths in the response
var RecordIDPaths = map[string]string{
	objectBoard: "ID",
	objectItem:  "ID",
	objectUser:  "ID",
}

// How to read & build these patterns: https://github.com/gobwas/glob
func supportedOperations() components.EndpointRegistryInput {
	// We support reading everything under schema.json, so we get all the objects and join it into a pattern.
	readSupport := []string{objectBoard, objectItem, objectUser}
	writeSupport := []string{objectBoard, objectItem, objectUser}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: objectItem,
				Support:  components.DeleteSupport,
			},
		},
	}
}

func getNextRecordsURL(_ *ajson.Node) (string, error) {
	// Pagination is not supported for this provider.
	return "", nil
}
