package providers

const Snowflake Provider = "snowflake"

func init() {
	SetInfo(Snowflake, ProviderInfo{
		DisplayName: "Snowflake",
		AuthType:    Custom,
		BaseURL:     "https://{{.workspace}}.snowflakecomputing.com",

		// Integration can only be installed over the API as of now.
		// Not using the UI library.
		CustomOpts: &CustomAuthOpts{},
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{},
			Regular:  &MediaTypeRegular{},
		},
	})
}
