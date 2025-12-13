package providers

const AcuityScheduling Provider = "acuityScheduling"

func init() {
	// AcuityScheduling Configuration
	SetInfo(AcuityScheduling, ProviderInfo{
		DisplayName: "Acuity Scheduling",
		AuthType:    Oauth2,
		BaseURL:     "https://acuityscheduling.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://acuityscheduling.com/oauth2/authorize",
			TokenURL:                  "https://acuityscheduling.com/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722403809/media/const%20AcuityScheduling%20Provider%20%3D%20%22acuityScheduling%22_1722403809.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722403830/media/const%20AcuityScheduling%20Provider%20%3D%20%22acuityScheduling%22_1722403830.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722403809/media/const%20AcuityScheduling%20Provider%20%3D%20%22acuityScheduling%22_1722403809.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722403830/media/const%20AcuityScheduling%20Provider%20%3D%20%22acuityScheduling%22_1722403830.svg",
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
