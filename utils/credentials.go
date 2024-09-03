package utils

import (
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common/scanning"
	"golang.org/x/oauth2"
)

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

func SalesforceOAuthConfigFromRegistry(registry scanning.Registry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)
	salesforceWorkspace := registry.MustString(WorkspaceRef)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", salesforceWorkspace),
			TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", salesforceWorkspace),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	return cfg
}

func SalesforceOauthTokenFromRegistry(registry scanning.Registry) *oauth2.Token {
	accessToken := registry.MustString(AccessToken)
	refreshToken := registry.MustString(RefreshToken)

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	return tok
}
