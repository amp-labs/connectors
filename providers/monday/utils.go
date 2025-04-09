package monday

import (
	"fmt"
	"strconv"
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

func makeNextRecordsURL(params common.ReadParams, count int) func(*ajson.Node) (string, error) {
	return func(node *ajson.Node) (string, error) {
		if count < defaultPageSize {
			return "", nil
		}

		var currentPage int
		if params.NextPage != "" {
			_, err := fmt.Sscanf(string(params.NextPage), "%d", &currentPage)
			if err != nil {
				return "", fmt.Errorf("invalid next page format: %w", err)
			}
		}

		nextPage := currentPage + count

		return strconv.Itoa(nextPage), nil
	}
}
