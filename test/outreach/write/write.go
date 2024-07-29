package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	outreach_test "github.com/amp-labs/connectors/test/outreach"
	"github.com/amp-labs/connectors/test/utils"
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

func main() {
	var err error
	ctx := context.Background()

	conn := outreach_test.GetOutreachConnector(context.Background(), DefaultCredsFile)

	// Set up slog logging.
	utils.SetupLogging()

	err = testWriteConnector(ctx, conn)
	if err != nil {
		slog.Error("Error testing", "connector", conn, "error", err)
	}
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
