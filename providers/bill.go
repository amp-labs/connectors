package providers

const Bill Provider = "bill"

func init() {
	SetInfo(Bill, ProviderInfo{
		DisplayName:        "Bill",
		AuthType:           Custom,
		BaseURL:            "https://gateway.prod.bill.com",
		PostAuthInfoNeeded: true,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765453519/BILL.D-4df6115f_lywcbk.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765453293/media/bill.com_1765453291.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765453335/media/bill.com_1765453335.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765453314/media/bill.com_1765453314.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "username",
					DisplayName: "Username",
					DocsURL:     "https://developer.bill.com/reference/login",
					Prompt:      "Enter your BILL account email",
				},
				{
					Name:        "password",
					DisplayName: "Password",
					DocsURL:     "https://developer.bill.com/reference/login",
					Prompt:      "Enter your BILL account password",
				},
				{
					Name:        "organizationId",
					DisplayName: "Organization Id",
					DocsURL:     "https://developer.bill.com/reference/login",
					Prompt:      "Enter your BILL Organization ID",
				},
				{
					Name:        "devKey",
					DisplayName: "Developer Key",
					DocsURL:     "https://developer.bill.com/reference/login",
					Prompt:      "Enter your BILL Developer Key",
				},
			},
		},
	})
}
