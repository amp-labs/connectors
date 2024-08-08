package providers

const (
	CustomerDataPipelines Provider = "customerDataPipelines"
	CustomerJourneysApp   Provider = "customerJourneysApp"
	CustomerJourneysTrack Provider = "customerJourneysTrack"
)

func init() { //nolint:funlen
	SetInfo(CustomerDataPipelines, ProviderInfo{
		DisplayName: "Customer.io Data Pipelines",
		AuthType:    Basic,
		BaseURL:     "https://cdp.customer.io/v1",
		BasicOpts: &BasicAuthOpts{
			DocsURL: "https://customer.io/docs/api/cdp/#section/Authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349524/media/customerDataPipelines_1722349524.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349458/media/customerDataPipelines_1722349458.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349524/media/customerDataPipelines_1722349524.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349482/media/customerDataPipelines_1722349482.svg",
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

	SetInfo(CustomerJourneysApp, ProviderInfo{
		AuthType:    ApiKey,
		DisplayName: "Customer.io Journeys App",
		BaseURL:     "https://api.customer.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://customer.io/docs/api/app/#section/Authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349524/media/customerDataPipelines_1722349524.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349458/media/customerDataPipelines_1722349458.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349524/media/customerDataPipelines_1722349524.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349482/media/customerDataPipelines_1722349482.svg",
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

	SetInfo(CustomerJourneysTrack, ProviderInfo{
		DisplayName: "Customer.io Journeys Track",
		AuthType:    Basic,
		BaseURL:     "https://track.customer.io",
		BasicOpts: &BasicAuthOpts{
			DocsURL: "https://customer.io/docs/api/track/#section/Authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349524/media/customerDataPipelines_1722349524.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349458/media/customerDataPipelines_1722349458.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349524/media/customerDataPipelines_1722349524.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349482/media/customerDataPipelines_1722349482.svg",
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
