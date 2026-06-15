package providers

import "net/http"

const Breezy Provider = "breezy"

func init() {
	SetInfo(Breezy, ProviderInfo{
		DisplayName: "Breezy HR",
		AuthType:    ApiKey,
		BaseURL:     "https://api.breezy.hr",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "",
			},
			DocsURL: "https://developer.breezy.hr/reference/authorization",
		},
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://api.breezy.hr/v3/companies",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "",
				LogoURL: "",
			},
			Regular: &MediaTypeRegular{
				IconURL: "",
				LogoURL: "",
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
