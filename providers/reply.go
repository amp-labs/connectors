package providers

import "net/http"

const Reply Provider = "reply"

func init() {
	// Reply (reply.io) API key authentication.
	// Keys are created in the Reply app under Settings > API keys and are sent
	// as a bearer token. Personal keys act on behalf of the user; team and
	// organization keys exist but require X-User-Id/X-Team-Id impersonation
	// headers on every request, so personal keys are the fit for connections.
	// Reply also runs an OAuth2 server (oauth.reply.io), but clients are
	// allow-listed by Reply and there is no self-serve OAuth app registration.
	// https://docs.reply.io/api-reference/authentication
	SetInfo(Reply, ProviderInfo{
		DisplayName: "Reply",
		AuthType:    ApiKey,
		BaseURL:     "https://api.reply.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.reply.io/api-reference/authentication",
		},
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://api.reply.io/v3/whoami",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1787152537/media/Catalog_1787152536.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1787152567/media/Catalog_1787152567.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1787152537/media/Catalog_1787152536.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1787152567/media/Catalog_1787152567.svg",
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
