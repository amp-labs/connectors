package providers

const Pipeliner Provider = "pipeliner"

func init() {
	// Pipeliner API Key authentication
	SetInfo(Pipeliner, ProviderInfo{
		// TODO [ExplicitWorkspaceRequired: true]
		DisplayName: "Pipeliner",
		AuthType:    Basic,
		BaseURL:     "https://eu-central.api.pipelinersales.com",
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169124/media/oli5civg2ug9a9tlpl3p.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409690/media/const%20Pipeliner%20Provider%20%3D%20%22pipeliner%22_1722409689.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409690/media/const%20Pipeliner%20Provider%20%3D%20%22pipeliner%22_1722409689.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409690/media/const%20Pipeliner%20Provider%20%3D%20%22pipeliner%22_1722409689.png",
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
	})
}
