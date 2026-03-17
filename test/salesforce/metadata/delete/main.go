package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	objectName = "Account"
	fieldName  = "amptestmetadatafield__c"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	// Step 1: Create the field via UpsertMetadata.
	slog.Info("Step 1: Creating field via UpsertMetadata", "object", objectName, "field", fieldName)

	upsertRes, err := conn.UpsertMetadata(ctx, &common.UpsertMetadataParams{
		Fields: map[string][]common.FieldDefinition{
			objectName: {
				{
					FieldName:   fieldName,
					DisplayName: "Amp Test Metadata Field",
					Description: "Temporary field created by integration test. Safe to delete.",
					ValueType:   common.FieldTypeBoolean,
					StringOptions: &common.StringFieldOptions{
						DefaultValue: goutils.Pointer("false"),
					},
				},
			},
		},
	})
	if err != nil {
		utils.Fail("error creating field", "error", err)
	}

	slog.Info("UpsertMetadata result:")
	utils.DumpJSON(upsertRes, os.Stdout)

	// Step 2: List object metadata to verify the field exists.
	slog.Info("Step 2: Listing metadata to verify field exists")
	printFieldMetadata(ctx, conn, objectName, fieldName)

	// Step 3: Delete the field via DeleteMetadata.
	slog.Info("Step 3: Deleting field via DeleteMetadata", "object", objectName, "field", fieldName)

	deleteRes, err := conn.DeleteMetadata(ctx, &common.DeleteMetadataParams{
		Fields: map[string][]string{
			objectName: {fieldName},
		},
	})
	if err != nil {
		utils.Fail("error deleting field", "error", err)
	}

	slog.Info("DeleteMetadata result:")
	utils.DumpJSON(deleteRes, os.Stdout)

	// Step 4: List object metadata again to verify the field is gone.
	slog.Info("Step 4: Listing metadata to verify field is deleted")
	printFieldMetadata(ctx, conn, objectName, fieldName)

	// Step 5: Try deleting the same field again to see what error Salesforce returns.
	slog.Info("Step 5: Deleting the same field again (should fail)")

	deleteRes2, err := conn.DeleteMetadata(ctx, &common.DeleteMetadataParams{
		Fields: map[string][]string{
			objectName: {fieldName},
		},
	})
	if err != nil {
		slog.Info("Second delete returned error (expected)", "error", err)
	} else {
		slog.Info("Second delete result (unexpected success):")
		utils.DumpJSON(deleteRes2, os.Stdout)
	}

	slog.Info("Done!")
}

type metadataLister interface {
	ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error)
}

func printFieldMetadata(ctx context.Context, conn metadataLister, object, field string) {
	metadata, err := conn.ListObjectMetadata(ctx, []string{object})
	if err != nil {
		utils.Fail("error listing metadata", "error", err)
	}

	objMeta, ok := metadata.Result[fmt.Sprintf("%s", object)]
	if !ok {
		// Try lowercase since Salesforce normalizes to lowercase.
		for key, val := range metadata.Result {
			objMeta = val
			_ = key

			break
		}
	}

	if fieldMeta, found := objMeta.Fields[field]; found {
		slog.Info("Field found", "field", field, "metadata", fieldMeta)
	} else {
		slog.Info("Field NOT found", "field", field)
	}
}
