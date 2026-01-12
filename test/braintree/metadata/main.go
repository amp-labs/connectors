package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/braintree"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := braintree.GetBraintreeConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"customers",
		"transactions",
		"disputes",
		"refunds",
		"verifications",
		"merchantAccounts",
		"payments",
		"inStoreLocations",
		"inStoreReaders",
		"businessAccountCreationRequests",
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
