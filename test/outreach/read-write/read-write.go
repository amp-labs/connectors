package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/outreach"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
	"github.com/brianvoe/gofakeit/v6"
)

const (
	DefaultCredsFile = "creds.json"
)

type Attribute struct {
	Email     string `json:"email"`
	EmailType string `json:"emailType"`
	Order     int    `json:"order"`
	Status    string `json:"status"`
}

type EmailAddress struct {
	Attributes Attribute `json:"attributes"`
	Type       string    `json:"type"`
}

type EmailAddressUpdate struct {
	Attributes Attribute `json:"attributes"`
	Type       string    `json:"type"`
	ID         int       `json:"id"` // necessary in updating
}

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
	var err error

	outreach := GetOutreachConnector(context.Background(), DefaultCredsFile)

	err = testReadConnector(context.Background(), outreach)
	if err != nil {
		log.Fatal(err)
	}

	err = testWriteConnector(context.Background(), outreach)
	if err != nil {
		log.Fatal(err)
	}
}

func testReadConnector(ctx context.Context, conn connectors.ReadConnector) error {
	config := connectors.ReadParams{
		ObjectName: "users",
		// NextPage:   "https://api.outreach.io/api/v2/users?page%5Blimit%5D=1\u0026page%5Boffset%5D=2",
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testWriteConnector(ctx context.Context, conn connectors.WriteConnector) error {
	var err error

	err = testCreate(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdate(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreate(ctx context.Context, conn connectors.WriteConnector) error {
	attribute := Attribute{
		Email:     gofakeit.Email(),
		EmailType: "email",
		Order:     0,
		Status:    "null",
	}

	config := common.WriteParams{
		ObjectName: "emailAddresses",
		RecordData: EmailAddress{
			Attributes: attribute,
			Type:       "emailAddress",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testUpdate(ctx context.Context, conn connectors.WriteConnector) error {
	patt := Attribute{
		Email: "groverstiedemann@lehner.io",
		Order: gofakeit.Number(0, 10),
	}

	config := common.WriteParams{
		ObjectName: "emailAddresses",
		RecordId:   "5",
		RecordData: EmailAddressUpdate{
			Attributes: patt,
			ID:         5,
			Type:       "emailAddress",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
