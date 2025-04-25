package providers

const AWSIdentityCenter Provider = "awsIdentityCenter"

func init() {
	SetInfo(AWSIdentityCenter, ProviderInfo{
		DisplayName: "AWS Identity Center",
		AuthType:    Basic,
		BaseURL:     "https://AWS_SERVICE_PLACEHOLDER.{{.region}}.amazonaws.com",
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
				IconURL: "",
				LogoURL: "",
			},
			Regular: &MediaTypeRegular{
				IconURL: "",
				LogoURL: "",
			},
		},
	})
}
