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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "company_id",
					DisplayName: "Company ID",
					DocsURL:     "https://developer.breezy.hr/reference/customer-guide",
					Prompt: "Copy the company `_id` from GET /v3/companies. " +
						"Required for company-scoped objects such as positions, pipelines, categories, " +
						"departments, questionnaires, and templates.",
				},
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1781523563/media/breezy.hr_1781523562.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1781284996/breezy_hr_logo_dark_jckpzm.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1781523563/media/breezy.hr_1781523562.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1781285012/breezy_hr_logo_light_gugdnv.png",
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
