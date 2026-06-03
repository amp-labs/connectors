package providers

import "github.com/amp-labs/connectors/common"

const (
	Meta Provider = "meta"

	// ModuleMetaWhatsApp is the module used for sending WhatsApp messages via the Cloud API.
	ModuleMetaWhatsApp common.ModuleID = "whatsapp"
)

func init() { //nolint:funlen
	SetInfo(Meta, ProviderInfo{
		DisplayName: "Meta",
		AuthType:    Oauth2,
		BaseURL:     "https://graph.facebook.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.facebook.com/v25.0/dialog/oauth",
			TokenURL:                  "https://graph.facebook.com/v25.0/oauth/access_token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753098801/media/meta.com_1753098801.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753098836/media/meta.com_1753098836.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753098801/media/meta.com_1753098801.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753098858/media/meta.com_1753098858.svg",
			},
		},
		DefaultModule: ModuleMetaWhatsApp,
		Modules: &Modules{
			ModuleMetaWhatsApp: {
				BaseURL:     "https://graph.facebook.com",
				DisplayName: "WhatsApp Business Platform",
				Support: Support{
					Proxy:     false,
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "whatsappAccountId",
					DisplayName: "WhatsApp Business Account ID",
					DocsURL:     "https://developers.facebook.com/docs/whatsapp/business-management-api/get-started",
					Prompt: "The WhatsApp Business Account ID (WABA ID), found in Meta Business Manager " +
						"under Business Settings, WhatsApp Accounts.",
					ModuleDependencies: &ModuleDependencies{
						ModuleMetaWhatsApp: ModuleDependency{},
					},
				},
				{
					Name:        "whatsappPhoneNumberId",
					DisplayName: "Phone Number ID",
					DocsURL:     "https://developers.facebook.com/docs/whatsapp/cloud-api/reference/phone-numbers",
					Prompt: "The Phone Number ID, found in Meta Business Manager " +
						"under WhatsApp Manager, Phone numbers. " +
						"This is the numeric ID, not the actual phone number.",
					ModuleDependencies: &ModuleDependencies{
						ModuleMetaWhatsApp: ModuleDependency{},
					},
				},
			},
		},
	})
}
