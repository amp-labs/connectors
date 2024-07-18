package utils

import (
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

//nolint:gochecknoglobals
var (
	AccessToken  = "accessToken"
	RefreshToken = "refreshToken"
	ClientId     = "clientId"
	ClientSecret = "clientSecret"
	WorkspaceRef = "workspaceRef"
	Provider     = "provider"
)

func SalesforceOAuthConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
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

func SalesforceOauthTokenFromRegistry(registry CredentialsRegistry) *oauth2.Token {
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

func HubspotOauthTokenFromRegistry(registry CredentialsRegistry) *oauth2.Token {
	accessToken := registry.MustString(AccessToken)
	refreshToken := registry.MustString(RefreshToken)

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // just pretend it's expired already, whatever, it'll fetch a new one.
	}

	return tok
}

func HubspotOAuthConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.hubspot.com/oauth/authorize",
			TokenURL:  "https://api.hubapi.com/oauth/v1/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"crm.objects.contacts.read",
			"crm.objects.contacts.write",
			"crm.objects.deals.read",
			"crm.objects.line_items.read",
			"oauth",
			"crm.objects.companies.read",
			"tickets",
		},
	}

	return cfg
}

func MSDynamics365CRMConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)
	workspace := registry.MustString(WorkspaceRef)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			fmt.Sprintf("https://%v.crm.dynamics.com/user_impersonation", workspace),
			"offline_access",
		},
	}

	return cfg
}

func MSDynamics365CRMTokenFromRegistry(registry CredentialsRegistry) *oauth2.Token {
	accessToken := registry.MustString(AccessToken)
	refreshToken := registry.MustString(RefreshToken)

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // just pretend it's expired already, whatever, it'll fetch a new one.
	}

	return tok
}

func SalesloftConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.salesloft.com/oauth/authorize",
			TokenURL:  "https://accounts.salesloft.com/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{},
	}

	return cfg
}

func SalesloftTokenFromRegistry(registry CredentialsRegistry) *oauth2.Token {
	accessToken := registry.MustString(AccessToken)
	refreshToken := registry.MustString(RefreshToken)

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // just pretend it's expired already, whatever, it'll fetch a new one.
	}

	return tok
}

func OutreachOAuthConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://api.outreach.io/oauth/authorize",
			TokenURL:  "https://api.outreach.io/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"users.all",
			"accounts.read",
			"calls.all",
			"events.all",
			"teams.all",
		},
	}

	return cfg
}

func OutreachOauthTokenFromRegistry(registry CredentialsRegistry) *oauth2.Token {
	accessToken := registry.MustString(AccessToken)
	refreshToken := registry.MustString(RefreshToken)

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // just pretend it's expired already, whatever, it'll fetch a new one.
	}

	return tok
}

func IntercomConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.intercom.com/oauth",
			TokenURL:  "https://api.intercom.io/auth/eagle/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{},
	}

	return cfg
}

func IntercomTokenFromRegistry(registry CredentialsRegistry) *oauth2.Token {
	accessToken := registry.MustString(AccessToken)

	tok := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "bearer",
	}

	return tok
}

func GongOAuthConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "https://dev-api.withampersand.com/callbacks/v1/oauth",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://app.gong.io/oauth2/authorize",
			TokenURL:  "https://app.gong.io/oauth2/generate-customer-token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"api:calls:read:basic",
			"api:users:read",
			"api:calls:create:basic",
			"api:calls:read:basic",
			"api:meetings:user:delete",
			"api:meetings:user:update",
			"api:logs:read",
			"api:meetings:user:create",
			"api:workspaces:read",
		},
	}

	return cfg
}

func GongOauthTokenFromRegistry(registry CredentialsRegistry) *oauth2.Token {
	accessToken := registry.MustString(AccessToken)
	refreshToken := registry.MustString(RefreshToken)

	expiry, _ := time.Parse(time.RFC822, "26 May 24 14:56 +0600")

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       expiry,
	}

	return tok
}

func ZendeskSupportConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)
	workspace := registry.MustString(WorkspaceRef)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("https://%v.zendesk.com", workspace),
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%v.zendesk.com/oauth/authorizations/new", workspace),
			TokenURL:  fmt.Sprintf("https://%v.zendesk.com/oauth/tokens", workspace),
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{
			"read",
			"write",
		},
	}

	return cfg
}

func ZendeskSupportTokenFromRegistry(registry CredentialsRegistry) *oauth2.Token {
	accessToken := registry.MustString(AccessToken)

	tok := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "bearer",
	}

	return tok
}
