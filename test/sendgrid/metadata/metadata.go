package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/sendgrid"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := sendgrid.GetSendGridConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"contacts",
		"lists",
		"segments",
		"templates",
		"bounces",
		"event_webhook_settings",
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
