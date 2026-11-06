package providers

const Databricks Provider = "databricks"

//nolint:funlen
func init() {
	SetInfo(Databricks, ProviderInfo{
		DisplayName: "Databricks",
		AuthType:    Oauth2,
		// The workspace is the full host, since the suffix depends on the cloud:
		// dbc-a1b2c3d4-e5f6.cloud.databricks.com, adb-1234567890123456.7.azuredatabricks.net,
		// 1234567890123456.7.gcp.databricks.com.
		BaseURL: "https://{{.workspace}}",
		Oauth2Opts: &Oauth2Opts{
			// OAuth machine-to-machine: each customer supplies the client ID and secret of a
			// service principal in their own workspace.
			// https://docs.databricks.com/aws/en/dev-tools/auth/oauth-m2m
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://{{.workspace}}/oidc/v1/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
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
					Prompt: "The full workspace host without the scheme " +
						"(e.g. `dbc-a1b2c3d4-e5f6.cloud.databricks.com` or `adb-1234567890123456.7.azuredatabricks.net`).",
				},
				{
					Name:        "httpPath",
					DisplayName: "SQL Warehouse HTTP Path",
					Prompt: "The HTTP path from the SQL warehouse's connection details " +
						"(e.g. `/sql/1.0/warehouses/1a2b3c4d5e6f7a8b`).",
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
