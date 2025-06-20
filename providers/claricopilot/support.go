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

//nolint:gochecknoglobals
var (
	objectNameCalls             = "calls"
	objectNameUsers             = "users"
	objectNameTopics            = "topics"
	objectNameScorecard         = "scorecard"
	objectNameScorecardTemplate = "scorecard-template"
	objectNameContact           = "contacts"
	objectNameDeal              = "deals"
	objectNameAccount           = "accounts"
)

var supportedObjectV2 = datautils.NewSet( //nolint:gochecknoglobals
	objectNameTopics,
)

//nolint:gochecknoglobals
var writeObjectMapping = datautils.NewDefaultMap(map[string]string{
	objectNameCalls:   "create-call",
	objectNameContact: "create-contact",
	objectNameDeal:    "create-deal",
	objectNameAccount: "create-account",
},
	func(objectName string) (fieldName string) {
		return objectName
	},
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		objectNameCalls, objectNameUsers, objectNameTopics, objectNameScorecard, objectNameScorecardTemplate,
	}

	writeSupport := []string{
		objectNameCalls, objectNameContact, objectNameDeal, objectNameAccount,
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
