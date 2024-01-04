package test

import (
	"log/slog"
	"os"

	"github.com/spyzhov/ajson"
)

type Creds struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Subdomain    string `json:"subdomain"`
	Provider     string `json:"provider"`
}

func GetCreds(path string) (*Creds, error) {
	if path == "" {
		path = "../creds.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Error opening creds.json", "error", err)
		return nil, err
	}

	credsMap, err := ajson.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	clientId, err := credsMap.JSONPath("$.providerApp.clientId")
	if err != nil {
		return nil, err
	}

	clientSecret, err := credsMap.JSONPath("$.providerApp.clientSecret")
	if err != nil {
		return nil, err
	}

	accessToken, err := credsMap.JSONPath("$.AccessToken.Token")
	if err != nil {
		return nil, err
	}

	refreshToken, err := credsMap.JSONPath("$.RefreshToken.Token")
	if err != nil {
		return nil, err
	}

	subdomain, err := credsMap.JSONPath("$.providerWorkspaceRef")
	if err != nil {
		return nil, err
	}

	provider, err := credsMap.JSONPath("$.providerApp.provider")
	if err != nil {
		return nil, err
	}

	creds := &Creds{
		ClientId:     clientId[0].MustString(),
		ClientSecret: clientSecret[0].MustString(),
		AccessToken:  accessToken[0].MustString(),
		RefreshToken: refreshToken[0].MustString(),
		Subdomain:    subdomain[0].MustString(),
		Provider:     provider[0].MustString(),
	}

	return creds, nil
}
