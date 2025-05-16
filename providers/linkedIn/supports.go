package linkedIn

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	supportWrite := []string{
		"adAccounts",
		"adTargetTemplates",
		"adPublisherRestrictions",
		"inMailContents",
		"conversationAds",
		"adLiftTests",
		"adExperiments",
		"conversions",
		"thirdPartyTrackingTags",
		"events",
		"insightTags",
		"conversionEvents",
		"adPageSets",
		"dmpSegments",
		"leadForms",
		"ugcPosts",
		"posts",
		"creatives",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportWrite, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
