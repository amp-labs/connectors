package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/dropbox"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetDropboxConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"shared_links"})
	if err != nil {
		log.Fatal("Error fetching metadata: ", err)
	}

	for objName, objMeta := range m.Result {
		log.Printf("   - %s: %d fields\n", objName, len(objMeta.Fields))
	}

	utils.DumpJSON(m, os.Stdout)
}
