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

var MSDynamics365CRMWorkspace = "org5bd08fdd" //nolint:gochecknoglobals

func MSDynamics365CRMConfigFromRegistry(registry CredentialsRegistry) *oauth2.Config {
	clientId := registry.MustString(ClientId)
	clientSecret := registry.MustString(ClientSecret)

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
			fmt.Sprintf("https://%v.crm.dynamics.com/user_impersonation", MSDynamics365CRMWorkspace),
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

var GongWorkspace = "us-49467" //nolint:gochecknoglobals

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
			"api:users:read",
			"api:calls:create:basic",
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

	expiry, _ := time.Parse(time.RFC822, "25 Apr 24 14:56 +0600")

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       expiry,
	}

	return tok
}
