package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/outreach"
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

	conn := outreach.GetOutreachConnector(context.Background(), DefaultCredsFile)

	err = testReadConnector(context.Background(), conn)
	if err != nil {
		log.Fatal(err)
	}
}

func testReadConnector(ctx context.Context, conn connectors.ReadConnector) error {
	config := connectors.ReadParams{
		ObjectName: "sequences",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     []string{"openCount", "description"},
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
