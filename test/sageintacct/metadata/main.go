package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/sageintacct"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := sageintacct.GetSageIntacctConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"company-config/user", "general-ledger/account", "contracts/contract"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
