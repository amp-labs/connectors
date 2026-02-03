package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := phoneburner.GetPhoneBurnerConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"contacts", "folders", "members", "voicemails", "phonenumber", "dialsession", "customfields"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
