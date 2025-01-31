package providers

const (
	SolarWindsServiceDeskUS Provider = "solarWindsServiceDeskUS"
	SolarWindsServiceDeskEU Provider = "solarWindsServiceDeskEU"
	SolarWindsServiceDeskAU Provider = "solarWindsServiceDeskAU"
)

//nolint:funlen
func init() {
	// SolarWindsServiceDesk US(region) configuration
	SetInfo(SolarWindsServiceDeskUS, ProviderInfo{
		DisplayName: "solarWindsServiceDeskUS",
		AuthType:    ApiKey,
		BaseURL:     "https://api.samanage.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "X-Samanage-Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://apidoc.samanage.com/",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322100/media/solarwinds.com_1738322099.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322158/media/solarwinds.com_1738322159.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322100/media/solarwinds.com_1738322099.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322158/media/solarwinds.com_1738322159.svg",
			},
		},
	})

	// SolarWindsServiceDesk EU(region) configuration
	SetInfo(SolarWindsServiceDeskEU, ProviderInfo{
		DisplayName: "solarWindsServiceDeskEU",
		AuthType:    ApiKey,
		BaseURL:     "https://apieu.samanage.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "X-Samanage-Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://apidoc.samanage.com/",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322100/media/solarwinds.com_1738322099.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322158/media/solarwinds.com_1738322159.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322100/media/solarwinds.com_1738322099.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322158/media/solarwinds.com_1738322159.svg",
			},
		},
	})

	// SolarWindsServiceDesk APJ(region) configuration
	SetInfo(SolarWindsServiceDeskAU, ProviderInfo{
		DisplayName: "solarWindsServiceDeskAU",
		AuthType:    ApiKey,
		BaseURL:     "https://apiau.samanage.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "X-Samanage-Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://apidoc.samanage.com/",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322100/media/solarwinds.com_1738322099.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322158/media/solarwinds.com_1738322159.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322100/media/solarwinds.com_1738322099.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322158/media/solarwinds.com_1738322159.svg",
			},
		},
	})
}
