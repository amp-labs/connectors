package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/sendgrid"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetSendGridConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"contacts",
		"lists",
		"segments",
		"singlesends",
		"templates",
		"field_definitions",
		"verified_senders",
		"senders",
		"bounces",
		"blocks",
		"spam_reports",
		"unsubscribes",
		"invalid_emails",
		"asm_groups",
		"categories",
		"subusers",
		"event_webhook_settings",
		"parse_webhook_settings",
	})
	if err != nil {
		log.Fatal("Error fetching metadata: ", err)
	}

	for objName, objMeta := range m.Result {
		log.Printf("   - %s: %d fields\n", objName, len(objMeta.Fields))
	}

	utils.DumpJSON(m, os.Stdout)
}
