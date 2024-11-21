package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "campaigns"

// we want to compare fields returned by read and schema properties provided by metadata methods
// they must match for all such objects
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetTheGetResponseConnector(ctx)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for GetResponse", "error", err)
	}

	fmt.Println("Read object using all fields from ListObjectMetadata")

	requestFields := datautils.Map[string, string](metadata.Result[objectName].FieldsMap).KeySet()

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     requestFields,
	})
	if err != nil {
		utils.Fail("error reading from GetResponse", "error", err)
	} else {
		if response.Rows == 0 {
			utils.Fail("expected to read at least one record", "error", err)
		}

		givenFields := datautils.Map[string, any](response.Data[0].Fields).KeySet()

		difference := givenFields.Diff(requestFields)
		if len(difference) != 0 {
			utils.Fail("connector read didn't match requested fields", "difference", difference)
		}
	}

	fmt.Println("==> success fields requested from ListObjectMetadata are all present in Read.")
}
