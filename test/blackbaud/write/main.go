package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/blackbaud"
	"github.com/amp-labs/connectors/test/blackbaud"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testBatches(ctx)
	if err != nil {
		return 1
	}

	err = testEmailAddresses(ctx)
	if err != nil {
		return 1
	}

	err = testVolunteers(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testBatches(ctx context.Context) error {
	conn := blackbaud.GetBlackbaudConnector(ctx)

	slog.Info("Creating the CRM Administration batches")

	writeParams := common.WriteParams{
		ObjectName: "crm-adnmg/batches",
		RecordData: map[string]any{
			"batch_template_id": "36450caa-4c57-4de0-80cb-690896abbea6",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	return nil
}

func testEmailAddresses(ctx context.Context) error {
	conn := blackbaud.GetBlackbaudConnector(ctx)

	slog.Info("Creating the Crm Constituent emailaddresses")

	writeParams := common.WriteParams{
		ObjectName: "crm-conmg/emailaddresses",
		RecordData: map[string]any{
			"constituent_id": "4C6F50CE-701E-4453-A4AC-097A3FEB2364",
			"email_address":  "sample@gmail.com",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the Crm Constituent emailaddresses")

	updateParams := common.WriteParams{
		ObjectName: "crm-conmg/emailaddresses",
		RecordData: map[string]any{
			"email_address_type": "Email",
		},
		RecordId: writeRes.RecordId,
	}

	res, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(res); err != nil {
		return err
	}

	return nil
}

func testVolunteers(ctx context.Context) error {
	conn := blackbaud.GetBlackbaudConnector(ctx)

	slog.Info("Creating the Crm volunteers")

	writeParams := common.WriteParams{
		ObjectName: "crm-volmg/volunteers",
		RecordData: map[string]any{
			"constituent_id": "4C6F50CE-701E-4453-A4AC-097A3FEB2364",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the Crm volunteers")

	updateParams := common.WriteParams{
		ObjectName: "crm-volmg/volunteers",
		RecordData: map[string]any{
			"emergency_contact_name": "Demo",
		},
		RecordId: writeRes.RecordId,
	}

	res, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(res); err != nil {
		return err
	}

	return nil
}

func Write(ctx context.Context, conn *ap.Connector, payload common.WriteParams) (*common.WriteResult, error) {
	res, err := conn.Write(ctx, payload)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// unmarshal the write response.
func constructResponse(res *common.WriteResult) error {
	jsonStr, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
