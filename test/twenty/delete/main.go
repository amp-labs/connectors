package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/twenty"
	"github.com/amp-labs/connectors/test/utils"
)

// Soft-deletes a record, e.g. go run ./test/twenty/delete -object vendorContracts -id <uuid>. Without -id it
// creates a record named "Twenty connector delete test" and soft-deletes it.
func main() {
	objectName := flag.String("object", "companies", "object name (namePlural)")
	recordID := flag.String("id", "", "record id")
	flag.Parse()

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetTwentyConnector(ctx)

	if *recordID == "" {
		created, err := conn.Write(ctx, common.WriteParams{
			ObjectName: *objectName,
			RecordData: map[string]any{"name": "Twenty connector delete test"},
		})
		if err != nil {
			utils.Fail("error creating the record to delete", "object", *objectName, "error", err)
		}

		*recordID = created.RecordId
	}

	res, err := conn.Delete(ctx, common.DeleteParams{ObjectName: *objectName, RecordId: *recordID})
	if err != nil {
		utils.Fail("error deleting from Twenty", "object", *objectName, "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
