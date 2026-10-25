package providers

const Clarify Provider = "clarify"

func init() {
	SetInfo(Clarify, ProviderInfo{
		DisplayName: "Clarify",
		AuthType:    ApiKey,
		// Every workspace endpoint is under /v1/workspaces/{slug}, e.g.
		// https://api.clarify.ai/v1/workspaces/{slug}/objects/person/resources.
		// https://developer.clarify.ai/docs/getting-started/key-concepts#workspace
		// The slug follows the version segment, so it is not templated here: that would pin /v1 for proxy
		// callers. The deep connector builds the path from the `workspace` metadata; proxy callers add
		// /v1/workspaces/<slug> to their own path.
		BaseURL: "https://api.clarify.ai",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "api-key ",
			},
			DocsURL: "https://developer.clarify.ai/docs/getting-started/authentication#api-keys",
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Workspace Slug",
					DocsURL:     "https://developer.clarify.ai/docs/getting-started/key-concepts#workspace",
					Prompt:      "Your workspace slug appears in your Clarify login URL and under the user avatar menu.",
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792813349/media/clarify_dark_icon.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792813350/media/clarify_dark_logo.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792813352/media/clarify_regular_icon.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792813354/media/clarify_regular_logo.svg",
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
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	})
}
