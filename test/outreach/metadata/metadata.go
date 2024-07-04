package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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

	conn, err := outreach.NewConnector(
		outreach.WithClient(ctx, http.DefaultClient, cfg, tok),
	)
	if err != nil {
		testUtils.Fail("error creating outreach connector", "error", err)
	}

	return conn
}

func main() {
	objects := []string{"emailAddresses", "users", "ladecima"}

	ctx := context.Background()

	conn := GetOutreachConnector(ctx, "creds.json")

	m, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("The Object-Metadata: %v\n", m.Result)

	fmt.Printf("The Errors: %v\n", m.Errors)
}
