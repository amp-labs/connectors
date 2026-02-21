package providers

const BigQuery Provider = "bigquery"

func init() {
	SetInfo(BigQuery, ProviderInfo{
		DisplayName: "BigQuery",
		AuthType:    Custom,
		BaseURL:     "https://bigquery.googleapis.com",

		// BigQuery uses custom authentication via service account.
		// The *bigquery.Client is passed via CustomAuthenticatedClient.
		CustomOpts: &CustomAuthOpts{},
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{},
			Regular:  &MediaTypeRegular{},
		},
	})
}
