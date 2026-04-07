package providers

const BigQuery Provider = "bigquery"

//nolint:funlen
func init() {
	SetInfo(BigQuery, ProviderInfo{
		DisplayName: "BigQuery",
		AuthType:    Custom,
		BaseURL:     "https://bigquery.googleapis.com",
		CustomOpts: &CustomAuthOpts{
			// No Headers or QueryParams — the server constructs pre-authenticated
			// BigQuery clients from the service account key, then passes them to
			// the connector via CustomAuthenticatedClient.
			Inputs: []CustomAuthInput{
				{
					Name:        "serviceAccountKey",
					DisplayName: "Service Account Key (Base64)",
					Prompt:      "Base64-encoded JSON key file for a GCP service account.", //nolint:lll
					DocsURL:     "https://docs.withampersand.com/customer-guides/bigquery#3-create-and-download-a-key",
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
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "projectId",
					DisplayName: "GCP Project ID",
					Prompt:      "The human-readable project ID (e.g. `my-project-id`), not the numeric project number.",
					DocsURL:     "https://docs.withampersand.com/customer-guides/bigquery#4-gather-your-connection-details",
				},
				{
					Name:        "dataset",
					DisplayName: "Dataset Name",
					Prompt:      "The BigQuery dataset to read from (e.g. `analytics`).",
					DocsURL:     "https://docs.withampersand.com/customer-guides/bigquery#4-gather-your-connection-details",
				},
				{
					Name:        "timestampColumn",
					DisplayName: "Timestamp Column",
					Prompt:      "A TIMESTAMP or DATETIME column for incremental reads (e.g. `updated_at`).", //nolint:lll
					DocsURL:     "https://docs.withampersand.com/customer-guides/bigquery#4-gather-your-connection-details",
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{},
			Regular:  &MediaTypeRegular{},
		},
	})
}
