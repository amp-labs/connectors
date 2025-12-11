package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/blackbaud"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := blackbaud.GetBlackbaudConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx,
		[]string{
			"crm-adnmg/currencies",
			"crm-evtmg/registrants",
			"crm-fndmg/fundraisingpurposes",
			"crm-mktmg/appeals",
			"crm-prsmg/prospects",
			"crm-revmg/payments",
			"crm-volmg/jobs",
		})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
