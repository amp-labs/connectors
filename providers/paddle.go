package providers

const (
	PaddleSandbox Provider = "paddleSandbox"
	Paddle        Provider = "paddle"
)

//nolint:funlen
func init() {
	SetInfo(PaddleSandbox, ProviderInfo{
		DisplayName: "Paddle Sandbox",
		AuthType:    ApiKey,
		BaseURL:     "https://sandbox-api.paddle.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://developer.paddle.com/api-reference/about/authentication",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407082/media/const%20%20%20%20Paddle%20%20%20Provider%20%3D%20%22paddle%22_1722407081.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407118/media/const%20%20%20%20Paddle%20%20%20Provider%20%3D%20%22paddle%22_1722407117.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407082/media/const%20%20%20%20Paddle%20%20%20Provider%20%3D%20%22paddle%22_1722407081.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407118/media/const%20%20%20%20Paddle%20%20%20Provider%20%3D%20%22paddle%22_1722407117.svg",
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
	SetInfo(Paddle, ProviderInfo{
		DisplayName: "Paddle",
		AuthType:    ApiKey,
		BaseURL:     "https://api.paddle.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://developer.paddle.com/api-reference/about/authentication",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407082/media/const%20%20%20%20Paddle%20%20%20Provider%20%3D%20%22paddle%22_1722407081.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407118/media/const%20%20%20%20Paddle%20%20%20Provider%20%3D%20%22paddle%22_1722407117.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407082/media/const%20%20%20%20Paddle%20%20%20Provider%20%3D%20%22paddle%22_1722407081.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722407118/media/const%20%20%20%20Paddle%20%20%20Provider%20%3D%20%22paddle%22_1722407117.svg",
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
