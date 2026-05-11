// Integration test for Connector.ApexTriggerExists.
//
// Flow:
//  1. Create a checkbox custom field on Lead (the apex trigger references it,
//     so the field must exist before the trigger compiles).
//  2. Build a CDC apex trigger zip and deploy it via Metadata API.
//  3. ApexTriggerExists should report true for the deployed trigger.
//  4. ApexTriggerExists should report false for a guaranteed-missing trigger
//     name (negative case so we don't get a false positive).
//  5. Build a destructive-changes zip and deploy it to remove the trigger.
//  6. ApexTriggerExists should report false post-delete.
//  7. Delete the checkbox field.
//
// On any failure the script aborts via utils.Fail (os.Exit(1)) and the
// trigger / field may be left behind in Salesforce; manual cleanup may be
// required before re-running.
package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	objectName        = "Lead"
	checkboxAPIName   = "AmpExistenceTest__c"
	checkboxDisplay   = "Amp Existence Test"
	fakeTriggerName   = "amp_existence_does_not_exist_trigger"
	deployPollPeriod  = 10 * time.Second
	deployPollTimeout = 5 * time.Minute
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	triggerName, err := salesforce.GenerateApexTriggerNameForCDC(objectName)
	if err != nil {
		utils.Fail("GenerateApexTriggerNameForCDC failed", "error", err)
	}

	// Step 1: Create the checkbox field the trigger code will reference.
	fmt.Printf("====== Creating prereq field %s.%s ======\n", objectName, checkboxAPIName)
	createCheckboxField(ctx, conn)

	// Step 2: Build & deploy the apex trigger.
	fmt.Printf("====== Deploying apex trigger %s ======\n", triggerName)
	deployTrigger(ctx, conn, triggerName)

	// Step 3: Existence check should return true.
	fmt.Println("====== ApexTriggerExists (expect true) ======")
	assertExists(ctx, conn, triggerName, true)

	// Step 4: Negative case — fake trigger should return false.
	fmt.Printf("====== ApexTriggerExists for fake %s (expect false) ======\n", fakeTriggerName)
	assertExists(ctx, conn, fakeTriggerName, false)

	// Step 5: Destructively remove the trigger.
	fmt.Printf("====== Destroying apex trigger %s ======\n", triggerName)
	destroyTrigger(ctx, conn, triggerName)

	// Step 6: Existence check should return false.
	fmt.Println("====== ApexTriggerExists post-delete (expect false) ======")
	assertExists(ctx, conn, triggerName, false)

	// Step 7: Clean up the checkbox field.
	fmt.Printf("====== Cleaning up prereq field %s.%s ======\n", objectName, checkboxAPIName)
	deleteCheckboxField(ctx, conn)

	fmt.Println("====== Done ======")
}

func createCheckboxField(ctx context.Context, conn *salesforce.Connector) {
	if _, err := conn.UpsertMetadata(ctx, &common.UpsertMetadataParams{
		Fields: map[string][]common.FieldDefinition{
			objectName: {
				{
					FieldName:   checkboxAPIName,
					DisplayName: checkboxDisplay,
					Description: "Prereq field for ApexTriggerExists integration test. Safe to delete.",
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
			objectName: {checkboxAPIName},
		},
	}); err != nil {
		utils.Fail("DeleteMetadata failed", "error", err)
	}
}

func deployTrigger(ctx context.Context, conn *salesforce.Connector, triggerName string) {
	zipData, err := salesforce.ConstructApexTriggerZipForCDC(salesforce.ApexTriggerParams{
		ObjectName:  objectName,
		TriggerName: triggerName,
		IndicatorField: common.FieldDefinition{
			FieldName: checkboxAPIName,
			ValueType: common.FieldTypeBoolean,
		},
		WatchFields: []string{"Email", "Phone"},
	}, checkboxAPIName)
	if err != nil {
		utils.Fail("ConstructApexTriggerZipForCDC failed", "error", err)
	}

	deployID, err := conn.DeployMetadataZip(ctx, zipData)
	if err != nil {
		utils.Fail("DeployMetadataZip failed", "error", err)
	}

	waitForDeploy(ctx, conn, deployID)
}

func destroyTrigger(ctx context.Context, conn *salesforce.Connector, triggerName string) {
	zipData, err := salesforce.ConstructDestructiveApexTriggerZip(triggerName)
	if err != nil {
		utils.Fail("ConstructDestructiveApexTriggerZip failed", "error", err)
	}

	deployID, err := conn.DeployMetadataZip(ctx, zipData)
	if err != nil {
		utils.Fail("DeployMetadataZip (destructive) failed", "error", err)
	}

	waitForDeploy(ctx, conn, deployID)
}

func waitForDeploy(ctx context.Context, conn *salesforce.Connector, deployID string) {
	deadline := time.Now().Add(deployPollTimeout)

	for {
		if time.Now().After(deadline) {
			utils.Fail("deploy did not finish in time", "deployID", deployID, "timeout", deployPollTimeout)
		}

		result, err := conn.CheckDeployStatus(ctx, deployID)
		if err != nil {
			utils.Fail("CheckDeployStatus failed", "deployID", deployID, "error", err)
		}

		if result.Done {
			if !result.Success {
				utils.Fail("deploy completed but failed",
					"deployID", deployID,
					"status", result.Status,
					"errorMessage", result.ErrorMessage,
					"failures", result.ComponentFailures,
				)
			}

			fmt.Printf("Deploy succeeded (deployID=%s)\n", deployID)

			return
		}

		select {
		case <-ctx.Done():
			utils.Fail("context cancelled while waiting for deploy", "deployID", deployID, "error", ctx.Err())
		case <-time.After(deployPollPeriod):
		}
	}
}

func assertExists(ctx context.Context, conn *salesforce.Connector, triggerName string, want bool) {
	got, err := conn.ApexTriggerExists(ctx, triggerName)
	if err != nil {
		utils.Fail("ApexTriggerExists returned error", "trigger", triggerName, "error", err)
	}

	if got != want {
		utils.Fail("ApexTriggerExists returned unexpected value",
			"trigger", triggerName, "want", want, "got", got)
	}

	fmt.Printf("OK: ApexTriggerExists(%s) = %t\n", triggerName, got)
}
