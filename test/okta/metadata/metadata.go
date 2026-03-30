package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/okta"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := okta.GetOktaConnector(ctx)

	meta, err := conn.ListObjectMetadata(ctx, []string{
		"users", "groups", "apps", "logs", "devices",
		"idps", "authorizationServers", "trustedOrigins", "zones",
		"brands", "domains", "authenticators", "policies", "eventHooks", "features",
	})
	if err != nil {
		log.Fatalf("ListObjectMetadata error: %v", err)
	}

	utils.DumpJSON(meta, os.Stdout)
}
