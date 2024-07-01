package catalog

const RingCentral Provider = "ringCentral"

func init() {
	// RingCentral configuration
	SetInfo(RingCentral, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://platform.ringcentral.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 PKCE,
			AuthURL:                   "https://platform.ringcentral.com/restapi/oauth/authorize",
			TokenURL:                  "https://platform.ringcentral.com/restapi/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "owner_id",
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
