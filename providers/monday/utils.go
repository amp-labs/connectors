package monday

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/spyzhov/ajson"
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

func makeNextRecordsURL(baseURL string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Monday.com uses GraphQL cursor-based pagination
		// Pagination is handled in the GraphQL query itself, not via URLs
		return baseURL, nil
	}
}
