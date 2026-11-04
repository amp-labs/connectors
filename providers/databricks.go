package providers

const Databricks Provider = "databricks"

func init() {
	SetInfo(Databricks, ProviderInfo{
		DisplayName: "Databricks",
		AuthType:    ApiKey,
		// e.g. dbc-a1b2c3d4-e5f6.cloud.databricks.com, adb-1234567890123456.7.azuredatabricks.net
		BaseURL: "https://{{.workspace}}",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.databricks.com/aws/en/dev-tools/auth/pat",
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
					Name:        "workspace",
					DisplayName: "Workspace Host",
					Prompt:      "The workspace host without the scheme (e.g. `dbc-a1b2c3d4-e5f6.cloud.databricks.com`).",
				},
				{
					Name:        "warehouseId",
					DisplayName: "SQL Warehouse ID",
					Prompt:      "The ID of the SQL warehouse that runs the queries, found in the warehouse's connection details.",
				},
				{
					Name:        "catalog",
					DisplayName: "Catalog",
					Prompt:      "The Unity Catalog catalog that holds the tables to read (e.g. `main`).",
				},
				{
					Name:        "schema",
					DisplayName: "Schema",
					Prompt:      "The schema inside the catalog that holds the tables to read (e.g. `openworks`).",
				},
				{
					Name:        "primaryKey",
					DisplayName: "Primary Key Column",
					Prompt:      "A column that uniquely identifies a row in every table (e.g. `id`).",
				},
				{
					Name:        "timestampColumn",
					DisplayName: "Timestamp Column",
					Prompt:      "A TIMESTAMP column that is updated whenever a row changes (e.g. `updated_at`).",
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{},
			Regular:  &MediaTypeRegular{},
		},
	})
}
