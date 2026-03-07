package providers

const Workday Provider = "workday"

func init() {
	SetInfo(Workday, ProviderInfo{
		DisplayName: "Workday",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.workday.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.workday.com/ccx/oauth2/{{.tenantName}}/authorize",
			TokenURL:                  "https://{{.workspace}}.workday.com/ccx/oauth2/{{.tenantName}}/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1771450356/media/workday.com_1771450354.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1771450230/media/workday.com_1771450224.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1771450356/media/workday.com_1771450354.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1771450331/media/workday.com_1771450328.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Workday Host",
					DocsURL:     "https://doc.workday.com/",
				},
				{
					Name:        "tenantName",
					DisplayName: "Tenant Name",
					DocsURL:     "https://doc.workday.com/",
				},
			},
		},
	})
}
