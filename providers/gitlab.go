package providers

const GitLab Provider = "gitlab"

func init() {
	SetInfo(GitLab, ProviderInfo{
		DisplayName: "GitLab",
		AuthType:    Oauth2,
		BaseURL:     "https://gitlab.com",
		CustomOpts: &CustomAuthOpts{
			// https://docs.gitlab.com/api/rest/authentication/#personalprojectgroup-access-tokens
			Headers: []CustomAuthHeader{
				{
					Name:          "PRIVATE-TOKEN",
					ValueTemplate: "{{ .token }}",
				},
			},
			Inputs: []CustomAuthInput{
				{
					Name:        "token",
					DisplayName: "Access Token",
					Prompt:      "This can be a personal, project, or group access token.",
					DocsURL:     "https://gitlab.com/-/user_settings/personal_access_tokens",
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734003317/media/GitLab_1734003316.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734003260/media/GitLab_1734003258.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734003317/media/GitLab_1734003316.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734003350/media/GitLab_1734003349.svg",
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
