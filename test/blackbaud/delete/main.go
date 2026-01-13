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

	err := testEvents(ctx)
	if err != nil {
		return 1
	}

	err = testVolunteers(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testEvents(ctx context.Context) error {
	conn := blackbaud.GetBlackbaudConnector(ctx)

	slog.Info("Deleting the Crm event")

	deleteParams := common.DeleteParams{
		ObjectName: "crm-evtmg/events",
		RecordId:   "48a81a28-c61d-4bce-8272-ada71d061c5e",
	}

	res, err := Delete(ctx, conn, deleteParams)
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

	slog.Info("Deleting the Crm volunteers")

	deleteParams := common.DeleteParams{
		ObjectName: "crm-volmg/volunteers",
		RecordId:   "4C6F50CE-701E-4453-A4AC-097A3FEB2364",
	}

	res, err := Delete(ctx, conn, deleteParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(res); err != nil {
		return err
	}

	return nil
}

func Delete(ctx context.Context, conn *ap.Connector, payload common.DeleteParams) (*common.DeleteResult, error) {
	res, err := conn.Delete(ctx, payload)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// unmarshal the delete response.
func constructResponse(res *common.DeleteResult) error {
	jsonStr, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
