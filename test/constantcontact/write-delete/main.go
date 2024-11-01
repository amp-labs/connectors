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
	"github.com/amp-labs/connectors/providers/constantcontact"
	msTest "github.com/amp-labs/connectors/test/constantcontact"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

type ContactPayload struct {
	EmailAddress ContactEmailAddress `json:"email_address"`
	FirstName    string              `json:"first_name"`
	LastName     string              `json:"last_name"`
	CreateSource string              `json:"create_source,omitempty"` // for CREATE operation
	UpdateSource string              `json:"update_source,omitempty"` // for UPDATE operation
}

type ContactEmailAddress struct {
	Address          string `json:"address"`
	PermissionToSend string `json:"permission_to_send"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := msTest.GetConstantContactConnector(ctx)
	defer utils.Close(conn)

	slog.Info("> TEST Create/Update/Delete Contact")
	slog.Info("Creating Contact")

	email := gofakeit.Email()
	contact := createContact(ctx, conn, &ContactPayload{
		EmailAddress: ContactEmailAddress{
			Address:          email,
			PermissionToSend: "implicit",
		},
		FirstName:    "Johnathan",
		LastName:     "Doe",
		CreateSource: "Account",
	})

	slog.Info("Updating description of an Contact")
	updateContact(ctx, conn, contact.RecordId, &ContactPayload{
		EmailAddress: ContactEmailAddress{
			Address:          email,
			PermissionToSend: "implicit",
		},
		FirstName:    "John",
		LastName:     "Doe",
		UpdateSource: "Account",
	})

	slog.Info("View that contact has changed accordingly")

	res := readContacts(ctx, conn)

	updatedContact := searchContacts(res, "contact_id", contact.RecordId)
	for k, v := range map[string]string{
		"first_name": "John",
		"last_name":  "Doe",
	} {
		if !compare(updatedContact[k], v) {
			utils.Fail("error updated properties do not match", k, v, updatedContact[k])
		}
	}

	slog.Info("Removing this Contact")
	removeContact(ctx, conn, contact.RecordId)
	slog.Info("> Successful test completion")
}

func searchContacts(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if compare(data.Fields[key], value) {
			return data.Raw
		}
	}

	utils.Fail("error finding article")

	return nil
}

func readContacts(ctx context.Context, conn *constantcontact.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contact_id", "first_name", "last_name"),
	})
	if err != nil {
		utils.Fail("error reading from ConstantContact", "error", err)
	}

	return res
}

func createContact(ctx context.Context, conn *constantcontact.Connector, payload *ContactPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to ConstantContact", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a Article")
	}

	return res
}

func updateContact(
	ctx context.Context, conn *constantcontact.Connector, contactID string, payload *ContactPayload,
) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to ConstantContact", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a Article")
	}

	return res
}

func removeContact(ctx context.Context, conn *constantcontact.Connector, contactID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "contacts",
		RecordId:   contactID,
	})
	if err != nil {
		utils.Fail("error deleting for ConstantContact", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a Article")
	}
}

func compare(field any, value string) bool {
	if len(value) == 0 && field == nil {
		return true
	}

	switch field.(type) {
	case float64:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return false
		}

		return num == field.(float64)
	}

	return fmt.Sprintf("%v", field) == value
}
