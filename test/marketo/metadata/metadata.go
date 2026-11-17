package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/marketo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := marketo.GetMarketoConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"channels", "emailTemplates", "files", "folders", "meandyou", "leads"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
