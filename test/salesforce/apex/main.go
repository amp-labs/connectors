package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	pollInterval = 10 * time.Second
	pollTimeout  = 120 * time.Second
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	objectName := "Lead"
	triggerName := salesforce.GenerateApexTriggerName(objectName)

	// Deploy the apex trigger.
	fmt.Println("====== Deploying Apex Trigger ======")

	zipData, err := salesforce.ConstructApexTriggerZip(salesforce.ApexTriggerParams{
		ObjectName:        objectName,
		TriggerName:       triggerName,
		CheckboxFieldName: "agentIdentifierChanged__c",
		WatchFields:       []string{"Email", "Phone"},
	})
	if err != nil {
		utils.Fail("failed to construct apex trigger zip", "error", err)
	}

	saveZip("deploy.zip", zipData)
	printZipContents("deploy", zipData)

	deployID := deploy(ctx, conn, zipData)
	waitForDeploy(ctx, conn, deployID)

	// Delete the apex trigger via destructive changes.
	fmt.Println("====== Deleting Apex Trigger ======")

	destructiveZip, err := salesforce.ConstructDestructiveApexTriggerZip(triggerName)
	if err != nil {
		utils.Fail("failed to construct destructive apex trigger zip", "error", err)
	}

	saveZip("destructive.zip", destructiveZip)
	printZipContents("destructive", destructiveZip)

	deleteDeployID := deploy(ctx, conn, destructiveZip)
	waitForDeploy(ctx, conn, deleteDeployID)

	fmt.Println("====== Done ======")
}

func deploy(ctx context.Context, conn *salesforce.Connector, zipData []byte) string {
	deployID, err := conn.DeployMetadataZip(ctx, zipData)
	if err != nil {
		utils.Fail("deploy failed", "error", err)
	}

	slog.Info("Deploy initiated", "deployID", deployID)

	return deployID
}

func saveZip(name string, data []byte) {
	dir := filepath.Join("test", "salesforce", "apex", "output")

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		utils.Fail("failed to create output directory", "error", err)
	}

	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, data, os.ModePerm); err != nil {
		utils.Fail("failed to write zip file", "error", err)
	}

	slog.Info("Saved zip", "path", path)
}

func printZipContents(label string, data []byte) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		utils.Fail("failed to read zip", "error", err)
	}

	fmt.Printf("\n--- %s zip contents ---\n", label)

	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			utils.Fail("failed to open zip entry", "file", file.Name, "error", err)
		}

		content, err := io.ReadAll(rc)
		rc.Close()

		if err != nil {
			utils.Fail("failed to read zip entry", "file", file.Name, "error", err)
		}

		fmt.Printf("\n[%s]\n%s\n", file.Name, string(content))
	}

	fmt.Printf("--- end %s ---\n\n", label)
}

func waitForDeploy(ctx context.Context, conn *salesforce.Connector, deployID string) {
	ctx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			utils.Fail("timed out waiting for deploy", "deployID", deployID)
		case <-time.After(pollInterval):
			result, err := conn.CheckDeployStatus(ctx, deployID)
			if err != nil {
				utils.Fail("failed to check deploy status", "error", err)
			}

			slog.Info("Deploy status", "deployID", deployID, "status", result.Status, "done", result.Done, "success", result.Success)

			if result.Done {
				if !result.Success {
					slog.Error("deploy failed", "deployID", deployID, "status", result.Status, "errorMessage", result.ErrorMessage)

					for _, cf := range result.ComponentFailures {
						slog.Error("component failure",
							"componentType", cf.ComponentType,
							"fullName", cf.FullName,
							"problem", cf.Problem,
							"problemType", cf.ProblemType,
						)
					}

					utils.Fail("deploy completed but failed", "deployID", deployID, "status", result.Status)
				}

				slog.Info("Deploy succeeded", "deployID", deployID)

				return
			}
		}
	}
}
