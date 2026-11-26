package providers

// SageIntacctClientCredentials is a twin of the Sage Intacct provider that uses
// the client credentials grant. The connector in providers/sageintacct is reused.
// See "Client credentials grant type" in
// https://developer.sage.com/intacct/docs/1/sage-intacct-rest-api/authorization-and-security/oauth2
const SageIntacctClientCredentials Provider = "sageIntacctClientCredentials"

func init() {
	SetInfo(SageIntacctClientCredentials, ProviderInfo{
		DisplayName: "Sage Intacct (Client Credentials)",
		AuthType:    Oauth2,
		BaseURL:     "https://api.intacct.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://api.intacct.com/ia/api/v1/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			// Sage Intacct identifies the user to act as by userId@companyId.
			// The token is top-level; a request can target an entity with the
			// X-IA-API-Param-Entity header.
			TokenReqBodyParams: map[string]string{
				"username": "{{.userId}}@{{.companyId}}",
			},
			DocsURL: "https://developer.sage.com/intacct/docs/1/sage-intacct-rest-api/authorization-and-security/oauth2#client-credentials-grant-type", //nolint:lll
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1755508463/media/sage.com_1755508463.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1755508463/media/sage.com_1755508463.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1755508405/media/sage.com_1755508404.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1755508405/media/sage.com_1755508404.svg",
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
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "companyId",
					DisplayName: "Company ID",
					Prompt:      "The Company ID you enter when you sign in to Sage Intacct.",
				},
				{
					Name:        "userId",
					DisplayName: "Web Services User ID",
					Prompt: "The User ID of the Web Services user for this integration, " +
						"exactly as it appears in Sage Intacct. It is case-sensitive.",
				},
			},
		},
	})
}
