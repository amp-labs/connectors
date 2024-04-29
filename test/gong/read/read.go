package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Cursors are not empty in case there are 100+ records, if less - empty
	// If empty, the first page is returned

	var config connectors.ReadParams

	objectName := "calls"

	switch objectName {
	case "users":
		config = connectors.ReadParams{
			ObjectName: objectName,
			Fields:     []string{"id", "emailAddress", "created"},
		}
	case "calls":
		config = connectors.ReadParams{
			ObjectName: objectName,
			Fields:     []string{"id", "duration", "participants"},
		}
	case "interaction":
		config = connectors.ReadParams{
			ObjectName: objectName,
			Fields:     []string{"id", "type", "timestamp"},
		}
	case "scorecards":
		config = connectors.ReadParams{
			ObjectName: objectName,
			Fields:     []string{"id", "score", "comments"},
		}
	case "day-by-day":
		config = connectors.ReadParams{
			ObjectName: objectName,
			Fields:     []string{"date", "totalCalls", "totalDuration"},
		}
	case "aggregate-by-period":
		config = connectors.ReadParams{
			ObjectName: objectName,
			Fields:     []string{"period", "totalCalls", "totalDuration"},
		}
	case "aggregate":
		config = connectors.ReadParams{
			ObjectName: objectName,
			Fields:     []string{"totalCalls", "totalDuration"},
		}
	default:
		fmt.Println("Invalid object name")
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
