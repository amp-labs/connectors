package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/outreach"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

const (
	DefaultCredsFile = "creds.json"
)

func GetOutreachConnector(ctx context.Context, filePath string) *outreach.Connector {
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

	cfg := utils.OutreachOAuthConfigFromRegistry(registry)
	tok := utils.OutreachOauthTokenFromRegistry(registry)

	conn, err := connectors.Outreach(
		outreach.WithClient(ctx, http.DefaultClient, cfg, tok),
	)
	if err != nil {
		testUtils.Fail("error creating outreach connector", "error", err)
	}

	return conn
}

func main() {
	outreach := GetOutreachConnector(context.TODO(), DefaultCredsFile)

	config := connectors.ReadParams{
		ObjectName: "users",
		// NextPage:   "https://api.outreach.io/api/v2/users?page%5Blimit%5D=1\u0026page%5Boffset%5D=2",
	}
	result, err := outreach.Read(context.TODO(), config)
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
