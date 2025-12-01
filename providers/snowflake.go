package providers

const Snowflake Provider = "snowflake"

func init() {
	SetInfo(Snowflake, ProviderInfo{
		DisplayName: "Snowflake",
		AuthType:    Custom,
		BaseURL:     "https://{{.workspace}}.snowflakecomputing.com",
		CustomOpts: &CustomAuthOpts{
			Inputs: []CustomAuthInput{
				{
					Name:        "username",
					DisplayName: "Username",
					Prompt:      "The Snowflake username for key-pair authentication",
				},
				{
					Name:        "privateKey",
					DisplayName: "Private Key",
					Prompt:      "RSA private key in PEM format",
					DocsURL:     "https://docs.snowflake.com/en/user-guide/key-pair-auth",
				},
				{
					Name:        "privateKeyPassphrase",
					DisplayName: "Private Key Passphrase (Optional)",
					Prompt:      "Enter the passphrase if the private key is encrypted",
				},
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
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Account Identifier",
					Prompt:      "Your Snowflake account identifier (e.g., `abc12345` if `abc12345.snowflakecomputing.com` is your account URL).",
					DocsURL:     "https://docs.snowflake.com/en/user-guide/admin-account-identifier",
				},
				{
					Name:        "query",
					DisplayName: "SQL Query",
					Prompt:      "The SQL query that defines the data you want to sync.",
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{},
			Regular:  &MediaTypeRegular{},
		},
	})
}
