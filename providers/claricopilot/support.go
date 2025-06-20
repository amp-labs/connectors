package claricopilot

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

func responseField(objectName string) string {
	switch objectName {
	case "scorecard":
		return "scorecards"
	case "scorecard-template":
		return "scoreCardTemplates"
	default:
		return objectName
	}
}

var supportedObjectV2 = datautils.NewSet( //nolint:gochecknoglobals
	"topics",
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"calls", "users", "topics", "scorecard", "scorecard-template",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
