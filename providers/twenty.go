package providers

const Twenty Provider = "twenty"

func init() {
	SetInfo(Twenty, ProviderInfo{
		DisplayName: "Twenty",
		AuthType:    ApiKey,
		// Twenty is self-hosted or on Twenty Cloud, so the API is on the instance's own domain:
		// https://{your-domain}/ for self-hosted, https://api.twenty.com/ for Cloud.
		// https://docs.twenty.com/developers/extend/api
		// The BaseURL is the instance root rather than /rest, so proxy callers reach /rest, /graphql and
		// /rest/metadata alike.
		BaseURL: "https://{{.workspace}}",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.twenty.com/developers/extend/api",
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Instance Domain",
					DocsURL:     "https://docs.twenty.com/developers/extend/api",
					Prompt: "Enter the domain of your Twenty instance, without https:// " +
						"(e.g. crm.yourcompany.com if self-hosted, api.twenty.com on Twenty Cloud).",
				},
			},
		},
		// Twenty's public mark, from its repository at the tag the connector was tested against.
		// Not yet hosted on Ampersand's media CDN.
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://raw.githubusercontent.com/twentyhq/twenty/twenty/v2.50.0/packages/twenty-docs/logo.svg",
				LogoURL: "https://raw.githubusercontent.com/twentyhq/twenty/twenty/v2.50.0/packages/twenty-docs/logo.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://raw.githubusercontent.com/twentyhq/twenty/twenty/v2.50.0/packages/twenty-docs/logo.svg",
				LogoURL: "https://raw.githubusercontent.com/twentyhq/twenty/twenty/v2.50.0/packages/twenty-docs/logo.svg",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
