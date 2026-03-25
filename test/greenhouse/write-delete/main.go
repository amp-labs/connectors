package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	gh "github.com/amp-labs/connectors/providers/greenhouse"
	connTest "github.com/amp-labs/connectors/test/greenhouse"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

// Greenhouse Harvest API v3: Write and Delete for applications and users.
// https://harvestdocs.greenhouse.io/docs/overview-and-philosophy

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGreenhouseConnector(ctx)

	if err := testApplicationsCRUD(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := testUsersWrite(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

// testApplicationsCRUD tests create, update, and delete for applications.
func testApplicationsCRUD(ctx context.Context, conn *gh.Connector) error {
	slog.Info("=== Applications: Create, Update, Delete ===")

	// Create a fresh candidate via direct API call.
	// Candidates are not in the connector schema, so we use the HTTP client directly.
	slog.Info("Creating candidate...")

	candidateID, err := createCandidate(ctx, conn)
	if err != nil {
		return fmt.Errorf("error creating candidate: %w", err)
	}

	slog.Info("Created candidate", "candidateId", candidateID)

	// Create an application for the candidate.
	slog.Info("Creating application...")

	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "applications",
		RecordData: map[string]any{
			"candidate_id": candidateID,
			"job_id":       4364850008, // Existing job in sandbox.
		},
	})
	if err != nil {
		return fmt.Errorf("error creating application: %w", err)
	}

	utils.DumpJSON(createRes, os.Stdout)

	// Update the application.
	slog.Info("Updating application...", "recordId", createRes.RecordId)

	updateRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "applications",
		RecordId:   createRes.RecordId,
		RecordData: map[string]any{
			"source_id": 4020580008, // Existing source "Aplaix" in sandbox.
		},
	})
	if err != nil {
		return fmt.Errorf("error updating application: %w", err)
	}

	utils.DumpJSON(updateRes, os.Stdout)

	// Delete the application.
	slog.Info("Deleting application...", "recordId", createRes.RecordId)

	deleteRes, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "applications",
		RecordId:   createRes.RecordId,
	})
	if err != nil {
		return fmt.Errorf("error deleting application: %w", err)
	}

	utils.DumpJSON(deleteRes, os.Stdout)

	slog.Info("Application deleted successfully")

	return nil
}

// testUsersWrite tests create and update for users.
// Note: users do not have a destroy scope, so no delete test.
func testUsersWrite(ctx context.Context, conn *gh.Connector) error {
	slog.Info("=== Users: Create, Update ===")

	email := gofakeit.Email()

	// Create a user.
	slog.Info("Creating user...", "email", email)

	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "users",
		RecordData: map[string]any{
			"first_name":    gofakeit.FirstName(),
			"last_name":     gofakeit.LastName(),
			"primary_email": email,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	utils.DumpJSON(createRes, os.Stdout)

	// Update the user.
	slog.Info("Updating user...", "recordId", createRes.RecordId)

	updateRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "users",
		RecordId:   createRes.RecordId,
		RecordData: map[string]any{
			"first_name": gofakeit.FirstName(),
		},
	})
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	utils.DumpJSON(updateRes, os.Stdout)

	slog.Info("User write test completed")

	return nil
}

// createCandidate creates a candidate via direct API call since candidates
// are not in the connector schema. Returns the candidate ID.
func createCandidate(ctx context.Context, conn *gh.Connector) (float64, error) {
	payload := map[string]any{
		"first_name": gofakeit.FirstName(),
		"last_name":  gofakeit.LastName(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://harvest.greenhouse.io/v3/candidates", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := conn.HTTPClient().Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, err
	}

	candidate, ok := result["candidate"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("unexpected response: %s", string(respBody))
	}

	id, ok := candidate["id"].(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected candidate id type")
	}

	return id, nil
}
