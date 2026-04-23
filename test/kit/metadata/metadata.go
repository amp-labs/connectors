package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/kit"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := kit.GetKitConnector(ctx)

	// nolint
	m, err := conn.ListObjectMetadata(ctx, []string{"broadcasts", "custom_fields", "forms", "subscribers", "tags", "email_templates", "purchases", "segments", "sequences", "webhooks"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
