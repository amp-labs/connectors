package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/phoneburner"
	connTest "github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetPhoneBurnerConnector(ctx)

	ownerID, ownerFirstName, err := getOwner(ctx, conn)
	if err != nil {
		slog.Error("Failed to read members (needed for owner_id)", "error", err)
		return 1
	}
	slog.Info("Using owner/member", "user_id", ownerID, "first_name", ownerFirstName)

	slog.Info("=== contacts (create -> update -> delete) ===")

	firstName := gofakeit.FirstName()
	lastName := fmt.Sprintf("amp-wd-%s", gofakeit.UUID())
	email := fmt.Sprintf("amp-wd-%s@example.com", gofakeit.UUID())
	phone := gofakeit.Numerify("602555####")

	updatedFirstName := gofakeit.FirstName()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		map[string]any{
			"owner_id":    ownerID,
			"email":       email,
			"first_name":  firstName,
			"last_name":   lastName,
			"phone":       phone,
			"phone_type":  1,
			"phone_label": "Amp wd test",
			"notes":       "Created by Ampersand write-delete integration test",
		},
		map[string]any{
			"first_name": updatedFirstName,
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("contact_user_id", "first_name", "last_name"),
			WaitBeforeSearch:    3 * time.Second,
			RecordIdentifierKey: "contact_user_id",
			UpdatedFields: map[string]string{
				"first_name": updatedFirstName,
			},
		},
	)

	slog.Info("PhoneBurner contacts write-delete test completed successfully")
	return 0
}

func getOwner(ctx context.Context, conn *phoneburner.Connector) (userID string, firstName string, err error) {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "members",
		Fields:     connectors.Fields("user_id", "first_name"),
		PageSize:   1,
	})
	if err != nil {
		return "", "", err
	}
	if len(res.Data) == 0 {
		return "", "", fmt.Errorf("no members returned")
	}

	id, _ := res.Data[0].Raw["user_id"].(string)
	if id == "" {
		return "", "", fmt.Errorf("members response missing user_id")
	}
	fn, _ := res.Data[0].Raw["first_name"].(string)
	return id, fn, nil
}
