package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesfinity"
	connTest "github.com/amp-labs/connectors/test/salesfinity"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	// A userId must be passed (via creds.json or USER_ID env var).
	// GET /v1/team returns a single team object with name and users;
	// See: https://docs.salesfinity.ai/api-reference/endpoint/get-team
	conn := connTest.GetConnector(ctx)
	userIdField := credscanning.Field{
		Name:      "userId",
		PathJSON:  "userId",
		SuffixENV: "USER_ID",
	}
	filePath := credscanning.LoadPath(providers.Salesfinity)
	reader := utils.MustCreateProvCredJSON(filePath, false, userIdField)
	userId := reader.Get(userIdField)

	slog.Info("TEST Create/Delete contact list")
	slog.Info("Creating contact list")

	createRes := createContactList(ctx, conn, userId)
	utils.DumpJSON(createRes, os.Stdout)
	slog.Info("Removing this contact list")
	deleteRes := deleteContactList(ctx, conn, createRes.RecordId)
	utils.DumpJSON(deleteRes, os.Stdout)
	slog.Info("Successful test completion")
}

func createContactList(ctx context.Context, conn *salesfinity.Connector, userID string) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contact-lists",
		RecordData: map[string]any{
			"name":    "Test Contact List",
			"user_id": userID,
			"contacts": []any{
				map[string]any{
					"first_name": "John",
					"last_name":  "Doe",
					"email":      "john.doe@example.com",
					"company":    "Example Corp",
					"title":      "Software Engineer",
					"phone_numbers": []any{
						map[string]any{
							"type":         "mobile",
							"number":       "5551234567",
							"country_code": "+1",
						},
					},
				},
			},
		},
	})
	if err != nil {
		utils.Fail("error creating contact list", "error", err)
	}
	if !res.Success {
		utils.Fail("failed to create contact list")
	}
	return res
}

// Currently the contact-list object does not support delete.
// When creating its also added on contact-lists/csv, this is to remove it from there.
func deleteContactList(ctx context.Context, conn *salesfinity.Connector, listID string) *common.DeleteResult {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "contact-lists/csv",
		RecordId:   listID,
	})
	if err != nil {
		utils.Fail("error deleting contact list", "error", err)
	}
	if !res.Success {
		utils.Fail("failed to delete contact list")
	}
	return res
}
