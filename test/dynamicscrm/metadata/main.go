package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	connTest "github.com/amp-labs/connectors/test/dynamicscrm"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "contacts"

// we want to compare fields returned by read and schema properties provided by metadata methods
// they must match for all such objects
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetMSDynamics365CRMConnector(ctx)
	defer utils.Close(conn)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for microsoft CRM", "error", err)
	}

	fmt.Println("Read object using all fields from ListObjectMetadata")

	requestFields := handy.Map[string, string](metadata.Result[objectName].FieldsMap).Keys()

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     requestFields,
	})
	if err != nil {
		utils.Fail("error reading from microsoft CRM", "error", err)
	} else {
		if response.Rows == 0 {
			utils.Fail("expected to read at least one record", "error", err)
		}

		givenFields := handy.Map[string, any](response.Data[0].Fields).Keys()

		difference := handy.NewSet(givenFields).Diff(handy.NewSet(requestFields))
		if len(difference) != 0 {
			utils.Fail("connector read didn't match requested fields", "difference", difference)
		}
	}

	fmt.Println("==> success fields requested from ListObjectMetadata are all present in Read.")
}
