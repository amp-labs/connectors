package monday

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/spyzhov/ajson"
)

// How to read & build these patterns: https://github.com/gobwas/glob
func supportedOperations() components.EndpointRegistryInput {
	// We support reading everything under schema.json, so we get all the objects and join it into a pattern.
	readSupport := []string{mondayObjectBoard, mondayObjectItems, mondayObjectUser, mondayObjectDocs}
	writeSupport := []string{mondayObjectBoard, mondayObjectItems, mondayObjectUser}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: mondayObjectItems,
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
