package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/amp-labs/connectors/gong"

	"github.com/amp-labs/connectors"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

const (
	DefaultCredsFile = "creds.json"
)

func GetGongConnector(ctx context.Context, filePath string) *gong.Connector {
	registry := utils.NewCredentialsRegistry()

	readers := []utils.Reader{
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['clientId']",
			CredKey:  "clientId",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['clientSecret']",
			CredKey:  "clientSecret",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['refreshToken']",
			CredKey:  "refreshToken",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['accessToken']",
			CredKey:  "accessToken",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['provider']",
			CredKey:  "provider",
		},
	}
	registry.AddReaders(readers...)

	cfg := utils.GongOAuthConfigFromRegistry(registry)
	tok := utils.GongOauthTokenFromRegistry(registry)

	conn, err := connectors.Gong(
		gong.WithClient(ctx, http.DefaultClient, cfg, tok),
		gong.WithWorkspace(utils.GongWorkspace),
	)
	if err != nil {
		testUtils.Fail("error creating gong connector", "error", err)
	}

	return conn
}

func main() {
	gong := GetGongConnector(context.Background(), DefaultCredsFile)

	gong.BaseURL = gong.BaseURL + "/v2"

	config := connectors.ReadParams{
		ObjectName: "calls",
		Fields:     []string{"title", "duration"},
		NextPage:   "eyJhbGciOiJIUzI1NiJ9.eyJjYWxsSWQiOjQ5NTM3MDc2MDE3NzYyMzgzNjAsInRvdGFsIjoxNzksInBhZ2VOdW1iZXIiOjAsInBhZ2VTaXplIjoxMDAsInRpbWUiOiIyMDIyLTA5LTEzVDA5OjMwOjAwWiIsImV4cCI6MTcxNDQwODE5MX0.cFcPhzRLCwOve4PGXMvo16E0SgrPfCdHYI2PPnBNTEs",
	}

	result, err := gong.Read(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")
}
