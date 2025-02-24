package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/netsuite"
	"github.com/amp-labs/connectors/test/utils"
)

// We want to compare fields returned by read and schema properties provided by metadata methods.
// Properties from read must all be present in schema definition.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetNetsuiteConnector(ctx)

	response, err := conn.ListObjectMetadata(ctx, []string{
		"account",
		"contact",
		"salesorder",
	})
	if err != nil {
		utils.Fail("error listing metadata for Netsuite", "error", err)
	}

	for obj, metadata := range response.Result {
		fmt.Printf("Metadata for %s: %+v\n\n\n", obj, metadata)
	}
}

// TODO: Implement this when read is done.
//
//	for _, object := range objects {
//		response, err := conn.Read(ctx, common.ReadParams{
//			ObjectName: object,
//		})
//		if err != nil {
//			utils.Fail("error reading from Netsuite", "error", err)
//		}
//		if response.Rows == 0 {
//			utils.Fail("expected to read at least one record", "error", err)
//		}
//		metadata, err := conn.ListObjectMetadata(ctx, []string{
//			object,
//		})
//		if err != nil {
//			utils.Fail("error listing metadata for Netsuite", "error", err)
//		}
//		slog.Info("Comparing")
//		mismatchErr := mockutils.ValidateReadConformsMetadata(object, response.Data[0].Raw, metadata)
//		if mismatchErr != nil {
//			utils.Fail("Failure: Schema and payload response have mismatching fields", "error", mismatchErr)
//		} else {
//			slog.Info("Success: Object metadata schema and endpoint response have the same fields")
//		}
//	}
