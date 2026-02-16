package providers

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

const Hubspot Provider = "hubspot"

const (
	// ModuleHubspotCRM is the module used for accessing standard CRM objects.
	ModuleHubspotCRM common.ModuleID = "crm"
)

func init() { //nolint:funlen
	// Hubspot configuration
	SetInfo(Hubspot, ProviderInfo{
		DisplayName: "HubSpot",
		AuthType:    Oauth2,
		BaseURL:     "https://api.hubapi.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://api.hubapi.com/integrations/v1/me",
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.hubspot.com/oauth/authorize",
			TokenURL:                  "https://api.hubapi.com/oauth/v1/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Delete:    true,
			Proxy:     true,
			Read:      true,
			Subscribe: true,
			Write:     true,
		},
		DefaultModule: ModuleHubspotCRM,
		Modules: &Modules{
			ModuleHubspotCRM: {
				BaseURL:     "https://api.hubapi.com/crm",
				DisplayName: "HubSpot CRM",
				Support: Support{
					BatchWrite: &BatchWriteSupport{
						Create: BatchWriteSupportConfig{
							DefaultRecordLimit: goutils.Pointer(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
						Update: BatchWriteSupportConfig{
							DefaultRecordLimit: goutils.Pointer(100), // nolint:mnd
							ObjectRecordLimits: nil,
							Supported:          true,
						},
					},
					Delete:    true,
					Read:      true,
					Subscribe: true,
					Write:     true,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722479285/media/hubspot_1722479284.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722479245/media/hubspot_1722479244.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722479285/media/hubspot_1722479284.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722479265/media/hubspot_1722479265.svg",
			},
		},
		PostAuthInfoNeeded: true,

		// IMPORTANT: The fetching of this metadata is added as a special case in the server,
		// because it requires the access token in the path, which is not really possible to
		// do with the current set up. If we can find a way to do this with the current interface,
		// we should remove the special case in the server, and define the GetPostAuthInfo method
		// as a method on the Connector struct.
		Metadata: &ProviderMetadata{
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "ownerId",
				},
			},
		},
	})
}
