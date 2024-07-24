package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/amp-labs/connectors/common/scanning"
	"github.com/amp-labs/connectors/outreach"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
)

const (
	DefaultCredsFile = "creds.json"
)

func GetOutreachConnector(ctx context.Context, filePath string) *outreach.Connector {
	registry := scanning.NewRegistry()

	readers := []scanning.Reader{
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$['clientId']",
			KeyName:  "clientId",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$['clientSecret']",
			KeyName:  "clientSecret",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$['refreshToken']",
			KeyName:  "refreshToken",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$['accessToken']",
			KeyName:  "accessToken",
		},
		&scanning.JSONReader{
			FilePath: filePath,
			JSONPath: "$['provider']",
			KeyName:  "provider",
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
