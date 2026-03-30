package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/okta"
	connTest "github.com/amp-labs/connectors/test/okta"
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

	// Test creating and updating a group
	groupID, err := testCreateGroup(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdateGroup(ctx, conn, groupID); err != nil {
		return err
	}

	// Test creating and updating a user
	userID, err := testCreateUser(ctx, conn)
	if err != nil {
		return err
	}

	if err := testUpdateUser(ctx, conn, userID); err != nil {
		return err
	}

	return nil
}

func testCreateGroup(ctx context.Context, conn *okta.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "groups",
		RecordData: map[string]any{
			"profile": map[string]any{
				"name":        fmt.Sprintf("Test Group %s", gofakeit.UUID()[:8]),
				"description": "Test group created by connector test",
			},
		},
	}

	slog.Info("Creating group...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create group: %w", err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.RecordId, nil
}

func testUpdateGroup(ctx context.Context, conn *okta.Connector, groupID string) error {
	params := common.WriteParams{
		ObjectName: "groups",
		RecordId:   groupID,
		RecordData: map[string]any{
			"profile": map[string]any{
				"name":        fmt.Sprintf("Updated Group %s", gofakeit.UUID()[:8]),
				"description": "Updated group description",
			},
		},
	}

	slog.Info("Updating group...", "groupID", groupID)

	res, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testCreateUser(ctx context.Context, conn *okta.Connector) (string, error) {
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

	slog.Info("Creating user...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.RecordId, nil
}

func testUpdateUser(ctx context.Context, conn *okta.Connector, userID string) error {
	params := common.WriteParams{
		ObjectName: "users",
		RecordId:   userID,
		RecordData: map[string]any{
			"profile": map[string]any{
				"firstName": gofakeit.FirstName(),
				"lastName":  gofakeit.LastName(),
			},
		},
	}

	slog.Info("Updating user...", "userID", userID)

	res, err := conn.Write(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
