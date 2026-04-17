package providers

const Procore Provider = "procore"

func init() {
	SetInfo(Procore, ProviderInfo{
		DisplayName: "Procore",
		AuthType:    Oauth2,
		BaseURL:     "https://www.procore.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://login.procore.com/oauth/authorize",
			TokenURL:                  "https://login.procore.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/q_auto/f_auto/v1776428797/Image_1776428788_0_hanjjk.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/q_auto/f_auto/v1776428963/Procore-Logo-Signature-Design-PNG_vk5wro.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/q_auto/f_auto/v1776428739/Procore-Logo-Signage-Design-PNG_ftf2os.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776428922/media/procore.com_1776428921.svg",
			},
		},
	})
}
