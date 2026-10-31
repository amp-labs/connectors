package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/twenty"
	"github.com/amp-labs/connectors/test/utils"
)

// Creates a record, or updates one when -id is given, from a JSON object. Without arguments it creates a
// company named "Twenty connector test record". E.g.
//
//	go run ./test/twenty/write -object vendorContracts -record '{"name":"VC-1","renewalnotes":"Renew in May"}'
//	go run ./test/twenty/write -object vendorContracts -id <uuid> -record '{"contractstatus":"SIGNED"}'
func main() {
	objectName := flag.String("object", "companies", "object name (namePlural)")
	recordID := flag.String("id", "", "record id to update; empty to create")
	recordJSON := flag.String("record", `{"name": "Twenty connector test record"}`, "record as a JSON object")
	flag.Parse()

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	var record map[string]any
	if err := json.Unmarshal([]byte(*recordJSON), &record); err != nil {
		utils.Fail("invalid -record", "error", err)
	}

	conn := connTest.GetTwentyConnector(ctx)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: *objectName,
		RecordId:   *recordID,
		RecordData: record,
	})
	if err != nil {
		utils.Fail("error writing to Twenty", "object", *objectName, "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
