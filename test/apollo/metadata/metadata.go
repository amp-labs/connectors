package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/apollo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := apollo.GetApolloConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"opportunities", "contact_stages", "email_accounts", "typed_custom_fields", "opportunity_stages", "users", "deals", "labels", "contacts", "accounts", "mimi"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
