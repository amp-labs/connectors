package solarwinds

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/spyzhov/ajson"
)

func getNextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := schemas.ObjectNames().GetList(common.ModuleRoot)

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
