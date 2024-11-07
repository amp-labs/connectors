package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/keap"
	connTest "github.com/amp-labs/connectors/test/keap"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/searcher"
	"github.com/brianvoe/gofakeit/v6"
)

var objectName = "contacts"

type ContactPayload struct {
	EmailAddresses []EmailAddress `json:"email_addresses"`
}

type EmailAddress struct {
	Email string `json:"email"`
	Field string `json:"field"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetKeapConnector(ctx)
	defer utils.Close(conn)

	slog.Info("> TEST Create/Update/Delete contacts")
	slog.Info("Creating contact")

	email := gofakeit.Email()
	createContact(ctx, conn, &ContactPayload{
		EmailAddresses: []EmailAddress{{
			Email: email,
			Field: "EMAIL1",
		}},
	})

	slog.Info("Reading contacts")

	res := readContacts(ctx, conn)

	slog.Info("Finding recently created contact")

	contact := searchContactByEmail(res, email)
	contactID := fmt.Sprintf("%v", contact["id"])

	slog.Info("Updating email of a contact")

	newEmail := gofakeit.Email()
	updateContact(ctx, conn, contactID, &ContactPayload{
		EmailAddresses: []EmailAddress{{
			Email: newEmail,
			Field: "EMAIL1",
		}},
	})

	slog.Info("View that contact has changed accordingly")

	res = readContacts(ctx, conn)
	idAsInt, _ := strconv.ParseInt(contactID, 10, 64)
	contact = searchContactByID(res, idAsInt)
	addresses := contact["email_addresses"].([]any)

	actual := addresses[0].(map[string]any)["email"].(string)
	if actual != newEmail {
		utils.Fail("email didn't change")
	}

	slog.Info("Removing this contact")
	removeContact(ctx, conn, contactID)
	slog.Info("> Successful test completion")
}

func searchContactByID(res *common.ReadResult, value int64) map[string]any {
	return searcher.Find(res, []searcher.Key{{
		Type: searcher.Integer,
		At:   "id",
	}}, value)
}

func searchContactByEmail(res *common.ReadResult, value string) map[string]any {
	return searcher.Find(res, []searcher.Key{{
		Type:  searcher.Array,
		At:    "email_addresses",
		Index: 0,
	}, {
		Type: searcher.String,
		At:   "email",
	},
	}, value)
}

func readContacts(ctx context.Context, conn *keap.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"id", "email_addresses",
		),
	})
	if err != nil {
		utils.Fail("error reading from Keap", "error", err)
	}

	return res
}

func createContact(ctx context.Context, conn *keap.Connector, payload *ContactPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Keap", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a contact")
	}
}

func updateContact(
	ctx context.Context, conn *keap.Connector, contactID string, payload *ContactPayload,
) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   contactID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Keap", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a contact")
	}

	return res
}

func removeContact(ctx context.Context, conn *keap.Connector, contactID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   contactID,
	})
	if err != nil {
		utils.Fail("error deleting for Keap", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a contact")
	}
}
