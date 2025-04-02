package monday

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
)

const (
	objectBoard = "boards"
	objectItem  = "items"
	objectUser  = "users"
)

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
