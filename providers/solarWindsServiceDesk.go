package providers

const SolarWindsServiceDeskUS Provider = "solarWindsServiceDeskUS"

// SolarWinds Service Desk has data centers in three regions: US, EU, and APJ, each with a different base URL.
//
//nolint:lll
func init() {
	// SolarWindsServiceDesk US configuration
	SetInfo(SolarWindsServiceDeskUS, ProviderInfo{
		DisplayName: "SolarWinds Service Desk US",
		AuthType:    ApiKey,
		BaseURL:     "https://api.samanage.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "X-Samanage-Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://documentation.solarwinds.com/en/success_center/swsd/content/completeguidetoswsd/token-authentication-for-api-integration.htm#link4",
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
