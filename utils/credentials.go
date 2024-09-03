package utils

import "github.com/amp-labs/connectors/common/scanning"

//nolint:gochecknoglobals
var (
	// TODO replace this with values from credsregistry/fields.go.

	AccessToken  = "accessToken"
	RefreshToken = "refreshToken"
	ClientId     = "clientId"
	ClientSecret = "clientSecret"
	WorkspaceRef = "workspaceRef"
	Provider     = "provider"
	ApiKey       = "apiKey"
)

func ApolloAPIKeyFromRegistry(registry scanning.Registry) string {
	apiKey := registry.MustString(ApiKey)

	return apiKey
}
