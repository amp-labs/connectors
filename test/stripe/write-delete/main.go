package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/stripe"
	connTest "github.com/amp-labs/connectors/test/stripe"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/brianvoe/gofakeit/v6"
)

var objectName = "customers"

type CustomerPayload struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetStripeConnector(ctx)

	slog.Info("> TEST Create/Update/Delete customer")
	slog.Info("Creating customer")

	name := gofakeit.AppName()
	email := gofakeit.Email()
	createCustomer(ctx, conn, &CustomerPayload{
		Name:  name,
		Email: email,
	})

	slog.Info("Reading customer")

	res := readCustomers(ctx, conn)

	slog.Info("Finding recently created customer")

	customer := searchCustomers(res, "name", name)
	customerID := fmt.Sprintf("%v", customer["id"])

	slog.Info("Updating customer name")

	newName := gofakeit.AppName()
	updateCustomer(ctx, conn, customerID, &CustomerPayload{
		Name:  newName,
		Email: email,
	})

	slog.Info("View that customer has changed accordingly")

	res = readCustomers(ctx, conn)

	customer = searchCustomers(res, "id", customerID)
	if customerName, ok := customer["name"].(string); !ok || customerName != newName {
		utils.Fail("error updated name doesn't match")
	}

	slog.Info("Removing this customer")
	removeCustomer(ctx, conn, customerID)
	slog.Info("> Successful test completion")
}

func searchCustomers(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding customer")

	return nil
}

func readCustomers(ctx context.Context, conn *stripe.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"id", "name", "email",
		),
	})
	if err != nil {
		utils.Fail("error reading from Stripe", "error", err)
	}

	return res
}

func createCustomer(ctx context.Context, conn *stripe.Connector, payload *CustomerPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Stripe", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a customer")
	}
}

func updateCustomer(ctx context.Context, conn *stripe.Connector, customerID string, payload *CustomerPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   customerID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Stripe", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a customer")
	}
}

func removeCustomer(ctx context.Context, conn *stripe.Connector, customerID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   customerID,
	})
	if err != nil {
		utils.Fail("error deleting for Stripe", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a customer")
	}
}
