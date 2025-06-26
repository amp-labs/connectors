package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/campaignmonitor"
	"github.com/amp-labs/connectors/test/campaignmonitor"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testClients(ctx)
	if err != nil {
		return 1
	}

	err = testPeople(ctx)
	if err != nil {
		return 1
	}

	err = testSuppress(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testClients(ctx context.Context) error {
	conn := campaignmonitor.GetCampaignMonitorConnector(ctx)

	slog.Info("Creating the clients")

	writeParams := common.WriteParams{
		ObjectName: "clients",
		RecordData: map[string]any{
			"CompanyName": "Demo",
			"Country":     "Australia",
			"TimeZone":    "(GMT+10:00) Canberra, Melbourne, Sydney",
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

func testPeople(ctx context.Context) error {
	conn := campaignmonitor.GetCampaignMonitorConnector(ctx)

	slog.Info("Creating the people")

	writeParams := common.WriteParams{
		ObjectName: "people",
		RecordData: map[string]any{
			"EmailAddress": "sample2@gmail.com",
			"Name":         "Sparrow2",
			"AccessLevel":  23,
			"Password":     "sample2Sparrow",
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

func testSuppress(ctx context.Context) error {
	conn := campaignmonitor.GetCampaignMonitorConnector(ctx)

	slog.Info("Creating the suppress")

	writeParams := common.WriteParams{
		ObjectName: "suppress",
		RecordData: map[string]any{
			"EmailAddresses": []string{"one@example.com", "two@example.com"},
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
