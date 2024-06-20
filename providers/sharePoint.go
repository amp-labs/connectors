package providers

const SharePoint Provider = "sharePoint"

func init() {
	SetInfo(SharePoint, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.sharepoint.com/_api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:              AuthorizationCode,
			AuthURL:                "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:               "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			ExplicitScopesRequired: true,
			// TODO: Switch to post-auth metadata collection
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
	})
}
