package providers

const CampaignMonitor Provider = "campaignMonitor"

func init() {
	// CampaignMonitor Configuration
	SetInfo(CampaignMonitor, ProviderInfo{
		DisplayName: "Campaign Monitor",
		AuthType:    Oauth2,
		BaseURL:     "https://api.createsend.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.createsend.com/oauth",
			TokenURL:                  "https://api.createsend.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722508734/media/const%20CampaignMonitor%20Provider%20%3D%20%22campaignMonitor%22_1722508735.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722508817/media/const%20CampaignMonitor%20Provider%20%3D%20%22campaignMonitor%22_1722508819.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722508734/media/const%20CampaignMonitor%20Provider%20%3D%20%22campaignMonitor%22_1722508735.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722508817/media/const%20CampaignMonitor%20Provider%20%3D%20%22campaignMonitor%22_1722508819.svg",
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
