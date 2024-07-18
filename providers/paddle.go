package providers

const (
	PaddleSandbox Provider = "paddleSandbox"
	Paddle        Provider = "paddle"
)

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
