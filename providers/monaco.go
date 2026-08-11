package providers

import "net/http"

const Monaco Provider = "monaco"

func init() {
	// Monaco API Key authentication.
	// Keys are created in the Monaco app under Settings > API Keys and are
	// prefixed with "mks_". They are sent as a bearer token.
	// https://docs.monaco.com/auth
	SetInfo(Monaco, ProviderInfo{
		DisplayName: "Monaco",
		AuthType:    ApiKey,
		BaseURL:     "https://api.monaco.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.monaco.com/auth",
		},
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://api.monaco.com/v1/me",
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
			Write:     false,
			// Monaco's list endpoints double as search endpoints, filtering on
			// a `filters` array. Only equality maps onto common.FilterOperator
			// today; Monaco itself also accepts contains/greater_than/less_than/is.
			Search: SearchSupport{
				Operators: SearchOperators{
					Equals: true,
				},
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1785823470/media/monaco.com_1785823468.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1785823514/media/monaco.com_1785823514.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1785823470/media/monaco.com_1785823468.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1785823514/media/monaco.com_1785823514.svg",
			},
		},
	})
}
