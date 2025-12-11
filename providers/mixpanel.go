package providers

const Mixpanel Provider = "mixpanel"

func init() {
	// Mixpanel configuration
	// apiserviceSubdomain ca either be [api, api-eu, data,data-eu].
	// Supported Mixpanel APIs
	// -	Ingestion API
	// -	Identity API
	// -	Event Export API
	// -	Data Pipelines API
	SetInfo(Mixpanel, ProviderInfo{
		DisplayName: "Mixpanel",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.mixpanel.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722597081/media/mixpanel.com_1722597079.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722597119/media/mixpanel.com_1722597118.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722597101/media/mixpanel.com_1722597100.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722597140/media/mixpanel.com_1722597139.svg",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		PostAuthInfoNeeded: false,
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "API Subdomain",
				},
			},
		},
	})
}
