package linkedin

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	supportRead := []string{
		"adTargetingFacets",
		"dmpEngagementSourceTypes",
		"adAccounts",
		"adCampaignGroups",
		"adCampaigns",
		"dmpSegments",
		"adAnalytics",
	}

	supportWrite := []string{
		"adAccounts",
		"adCampaignGroups",
		"adCampaigns",
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
		"adPageSets",
		"dmpSegments",
		"leadForms",
		"posts",
		"creatives",
	}

	supportDelete := []string{
		"adAccounts",
		"adCampaignGroups",
		"adTargetTemplates",
		"creatives",
		"adLiftTests",
		"thirdPartyTrackingTags",
		"events",
		"posts",
		"dmpSegments",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportRead, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportWrite, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportDelete, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
