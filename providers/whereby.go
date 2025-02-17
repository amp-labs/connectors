package providers

const Whereby Provider = "whereby"

//nolint:lll
func init() {
	// Whereby configuration
	SetInfo(Whereby, ProviderInfo{
		DisplayName: "Whereby",
		AuthType:    ApiKey,
		BaseURL:     "https://api.whereby.dev",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.whereby.com/whereby-101/readme/get-started-in-3-steps#step-1-generate-an-api-key",
		}, Support: Support{
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
		PostAuthInfoNeeded: false,
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739804309/media/whereby.com_1739804308.png",
				LogoURL: "https://framerusercontent.com/modules/4vhu5auio1F3btB69Kmz/OapwzfVmkVCDhgRB42g1/assets/byWcaCNWWZsZnXvSlFAudEUbl0.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739804309/media/whereby.com_1739804308.png",
				LogoURL: "https://framerusercontent.com/modules/4vhu5auio1F3btB69Kmz/OapwzfVmkVCDhgRB42g1/assets/byWcaCNWWZsZnXvSlFAudEUbl0.svg",
			},
		},
	})
}
