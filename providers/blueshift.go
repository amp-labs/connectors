package providers

const (
	BlueshiftEU Provider = "blueshiftEU"
	Blueshift   Provider = "blueshift"
)

func init() {
	// Blueshift configuration
	SetInfo(Blueshift, ProviderInfo{
		DisplayName: "Blueshift",
		AuthType:    Basic,
		BaseURL:     "https://api.getblueshift.com/api",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324992/media/blueshift_1722324992.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325053/media/blueshift_1722325053.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324964/media/blueshift_1722324964.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325021/media/blueshift_1722325020.svg",
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
		PostAuthInfoNeeded: false,
	})

	// BlueshiftEU connfiguration
	SetInfo(BlueshiftEU, ProviderInfo{
		DisplayName: "Blueshift (EU)",
		AuthType:    Basic,
		BaseURL:     "https://api.eu.getblueshift.com/api",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324992/media/blueshift_1722324992.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325053/media/blueshift_1722325053.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324964/media/blueshift_1722324964.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325021/media/blueshift_1722325020.svg",
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
		PostAuthInfoNeeded: false,
	})
}
