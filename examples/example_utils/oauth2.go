package example_utils

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

type OAuth2AuthCodeOptions struct {
	OAuth2ClientId     string
	OAuth2ClientSecret string
	OAuth2AccessToken  string
	OAuth2RefreshToken string
	Expiry             time.Time
}

func CreateOAuth2AuthorizationCodeClient(info *providers.ProviderInfo, opts OAuth2AuthCodeOptions) common.AuthenticatedHTTPClient {
	// Create the authenticated HTTP client.
	httpClient, err := info.NewClient(context.Background(), &providers.NewClientParams{
		// If you set this to true, the client will log all requests and responses.
		// Be careful with this in production, as it may expose sensitive data.
		Debug: *debug,

		// If you have your own HTTP client, you can use it here.
		Client: http.DefaultClient,

		OAuth2AuthCodeCreds: &providers.OAuth2AuthCodeParams{
			// Config represents the OAuth2 application's configuration. This is all known before the user authenticates.
			Config: &oauth2.Config{
				ClientID:     opts.OAuth2ClientId,
				ClientSecret: opts.OAuth2ClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:   info.Oauth2Opts.AuthURL,
					TokenURL:  info.Oauth2Opts.TokenURL,
					AuthStyle: oauth2.AuthStyleInParams,
				},
			},

			// Token represents the OAuth2 token. This is obtained after the user authenticates with a browser.
			// See scripts/oauth/token.go for an example of how to obtain this token (if you don't have one yet).
			Token: &oauth2.Token{
				AccessToken:  opts.OAuth2AccessToken,
				RefreshToken: opts.OAuth2RefreshToken,
				TokenType:    "bearer",
				Expiry:       opts.Expiry,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return httpClient
}
