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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1780944000/media/breezy.hr_1780944000.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1780944000/media/breezy.hr_1780944000.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1780944000/media/breezy.hr_1780944000.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1780944000/media/breezy.hr_1780944000.svg",
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
	})
}
