// Integration test for Connector.CustomFieldExists.
//
// Flow:
//  1. Create a checkbox custom field on Account via UpsertMetadata.
//  2. CustomFieldExists should report true for the new field.
//  3. CustomFieldExists should report false for a guaranteed-missing
//     field name (negative case so we don't get a false positive that returns
//     true for everything).
//  4. Delete the field via DeleteMetadata.
//  5. CustomFieldExists should report false post-delete.
//
// On any failure the script aborts via utils.Fail (os.Exit(1)) and the field
// may be left behind in Salesforce; re-running is idempotent because
// UpsertMetadata is idempotent and a residual field will be cleaned up by the
// test's own delete step.
package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	objectName       = "Account"
	fieldDisplayName = "Amp Existence Test"
	fieldAPIName     = "AmpExistenceTest__c"
	fakeFieldAPIName = "amp_existence_does_not_exist__c"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	// Step 1: Create the field.
	fmt.Printf("====== Creating %s.%s ======\n", objectName, fieldAPIName)
	createCheckboxField(ctx, conn)

	// Step 2: Existence check should return true.
	fmt.Println("====== CustomFieldExists (expect true) ======")
	assertExists(ctx, conn, objectName, fieldAPIName, true)

	// Step 3: Negative case — a fake field should return false.
	fmt.Printf("====== CustomFieldExists for fake %s (expect false) ======\n", fakeFieldAPIName)
	assertExists(ctx, conn, objectName, fakeFieldAPIName, false)

	// Step 4: Delete the field.
	fmt.Printf("====== Deleting %s.%s ======\n", objectName, fieldAPIName)
	deleteCheckboxField(ctx, conn)

	// Step 5: Existence check should return false.
	// Salesforce's Tooling API is read-from-cache for some queries; if this
	// flake bites the test, retry with backoff. For a freshly-deleted custom
	// field the SOQL query hits authoritative metadata so it's typically
	// immediate.
	fmt.Println("====== CustomFieldExists post-delete (expect false) ======")
	assertExists(ctx, conn, objectName, fieldAPIName, false)

	fmt.Println("====== Done ======")
}

func createCheckboxField(ctx context.Context, conn *salesforce.Connector) {
	if _, err := conn.UpsertMetadata(ctx, &common.UpsertMetadataParams{
		Fields: map[string][]common.FieldDefinition{
			objectName: {
				{
					FieldName:   fieldAPIName,
					DisplayName: fieldDisplayName,
					Description: "Temporary field created by CustomFieldExists integration test. Safe to delete.",
					ValueType:   common.FieldTypeBoolean,
					StringOptions: &common.StringFieldOptions{
						DefaultValue: new("false"),
					},
				},
			},
		},
	}); err != nil {
		utils.Fail("UpsertMetadata failed", "error", err)
	}
}

func deleteCheckboxField(ctx context.Context, conn *salesforce.Connector) {
	if _, err := conn.DeleteMetadata(ctx, &common.DeleteMetadataParams{
		Fields: map[common.ObjectName][]string{
			objectName: {fieldAPIName},
		},
	}); err != nil {
		utils.Fail("DeleteMetadata failed", "error", err)
	}
}

func assertExists(ctx context.Context, conn *salesforce.Connector, obj, field string, want bool) {
	got, err := conn.CustomFieldExists(ctx, obj, field)
	if err != nil {
		utils.Fail("CustomFieldExists returned error", "object", obj, "field", field, "error", err)
	}

	if got != want {
		utils.Fail("CustomFieldExists returned unexpected value",
			"object", obj, "field", field, "want", want, "got", got)
	}

	fmt.Printf("OK: CustomFieldExists(%s.%s) = %t\n", obj, field, got)
}
