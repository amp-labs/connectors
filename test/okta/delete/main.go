package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/okta"
	connTest "github.com/amp-labs/connectors/test/okta"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := connTest.GetOktaConnector(ctx)

	// Test delete flow: create a group, then delete it
	groupID, err := createTestGroup(ctx, conn)
	if err != nil {
		return err
	}

	if err := testDeleteGroup(ctx, conn, groupID); err != nil {
		return err
	}

	// Test user delete flow: create a user, deactivate (first delete), then permanently delete (second delete)
	// Note: Okta requires users to be deprovisioned before permanent deletion
	userID, err := createTestUser(ctx, conn)
	if err != nil {
		return err
	}

	// First delete deactivates the user (sets status to DEPROVISIONED)
	if err := testDeleteUser(ctx, conn, userID, "deactivate"); err != nil {
		return err
	}

	// Second delete permanently removes the user
	if err := testDeleteUser(ctx, conn, userID, "permanent"); err != nil {
		return err
	}

	return nil
}

func createTestGroup(ctx context.Context, conn *okta.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "groups",
		RecordData: map[string]any{
			"profile": map[string]any{
				"name":        fmt.Sprintf("Delete Test Group %s", gofakeit.UUID()[:8]),
				"description": "Test group to be deleted",
			},
		},
	}

	slog.Info("Creating test group for delete...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create group: %w", err)
	}

	slog.Info("Created group", "groupID", res.RecordId)

	return res.RecordId, nil
}

func testDeleteGroup(ctx context.Context, conn *okta.Connector, groupID string) error {
	params := common.DeleteParams{
		ObjectName: "groups",
		RecordId:   groupID,
	}

	slog.Info("Deleting group...", "groupID", groupID)

	res, err := conn.Delete(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Group deleted successfully")

	return nil
}

func createTestUser(ctx context.Context, conn *okta.Connector) (string, error) {
	email := gofakeit.Email()

	params := common.WriteParams{
		ObjectName: "users",
		RecordData: map[string]any{
			"profile": map[string]any{
				"firstName": gofakeit.FirstName(),
				"lastName":  gofakeit.LastName(),
				"email":     email,
				"login":     email,
			},
		},
	}

	slog.Info("Creating test user for delete...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	slog.Info("Created user", "userID", res.RecordId)

	return res.RecordId, nil
}

func testDeleteUser(ctx context.Context, conn *okta.Connector, userID string, stage string) error {
	params := common.DeleteParams{
		ObjectName: "users",
		RecordId:   userID,
	}

	slog.Info("Deleting user...", "userID", userID, "stage", stage)

	res, err := conn.Delete(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to delete user (%s): %w", stage, err)
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("User delete successful", "stage", stage)

	return nil
}
