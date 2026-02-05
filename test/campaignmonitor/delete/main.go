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

	return 0
}

func testClients(ctx context.Context) error {
	conn := campaignmonitor.GetCampaignMonitorConnector(ctx)

	slog.Info("Deleting the client")

	deleteParams := common.DeleteParams{
		ObjectName: "clients",
		RecordId:   "f676fc028bb37d27c0cdcf55e12a2069",
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
