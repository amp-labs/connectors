package providers

const Workday Provider = "workday"

//nolint:funlen
func init() {
	SetInfo(Workday, ProviderInfo{
		DisplayName: "Workday",
		AuthType:    Oauth2,
		// E.g. wd2-impl-services1.workday.com, wd3-services1.myworkday.com, etc.
		BaseURL: "https://{{.workspace}}",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			// NB: The authorization endpoint lives on the tenant's UI host (e.g. impl.workday.com),
			// not on the web services host, so it can't reuse the workspace like TokenURL does.
			AuthURL:                   "https://{{.authHost}}/{{.tenantName}}/authorize",
			TokenURL:                  "https://{{.workspace}}/ccx/oauth2/{{.tenantName}}/token",
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
			Proxy:     true,
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
					DisplayName: "Web Services Host",
					Prompt: "Host of the Workday REST API Endpoint shown in View API Clients, " +
						"e.g. `wd2-impl-services1.workday.com`.",
					DocsURL: "https://doc.workday.com/",
				},
				{
					Name:        "authHost",
					DisplayName: "Authorization Host",
					Prompt: "Host of the Authorization Endpoint shown in View API Clients, " +
						"e.g. `impl.workday.com` or `wd3.myworkday.com`.",
					DefaultValue: "impl.workday.com",
					DocsURL:      "https://doc.workday.com/",
				},
			},
		},
		ProviderAppMetadata: &ProviderAppMetadata{
			ProviderParams: []MetadataItemInput{
				{
					// The API client is registered inside the Workday tenant, so the
					// tenant is fixed per provider app and is needed to build the
					// OAuth URLs before any connection exists.
					Name:        "tenantName",
					DisplayName: "Tenant Name",
					Prompt: "Tenant name from the Token Endpoint shown in View API Clients. " +
						"It appears right after `/oauth2/` and before `/token`.",
					DocsURL: "https://doc.workday.com/",
				},
			},
		},
	})
}
