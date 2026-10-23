package providers

const Circleback Provider = "circleback"

func init() {
	SetInfo(Circleback, ProviderInfo{
		DisplayName: "Circleback",
		AuthType:    ApiKey,
		// The OpenAPI spec's server, https://circleback.ai/api. The API is not versioned.
		BaseURL: "https://circleback.ai/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://circleback.ai/docs/api#get-an-api-key",
		},
		Media: &Media{
			Regular: &MediaTypeRegular{
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792743779/media/circleback_regular_logo.svg",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
