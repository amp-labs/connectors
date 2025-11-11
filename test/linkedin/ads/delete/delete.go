package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/linkedin"
	"github.com/amp-labs/connectors/test/linkedin"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := TestAdCampaignsGroup(ctx)
	if err != nil {
		return 1
	}

	err = testAdAccounts(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testAdAccounts(ctx context.Context) error {
	conn := linkedin.GetAdsConnector(ctx)

	slog.Info("Deleting the adAccounts")

	deleteParams := common.DeleteParams{
		ObjectName: "adAccounts",
		RecordId:   "517370155",
	}

	res, err := Delete(ctx, conn, deleteParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func TestAdCampaignsGroup(ctx context.Context) error {
	conn := linkedin.GetAdsConnector(ctx)

	slog.Info("Deleting the adCampaignGroups")

	deleteParams := common.DeleteParams{
		ObjectName: "adCampaignGroups",
		RecordId:   "783551216",
	}

	res, err := Delete(ctx, conn, deleteParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func Delete(ctx context.Context, conn *ap.Connector, payload common.DeleteParams) (*common.DeleteResult, error) {
	res, err := conn.Delete(ctx, payload)
	if err != nil {
		return nil, err
	}

	return res, nil
}
