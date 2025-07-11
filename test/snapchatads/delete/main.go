package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/snapchatads"
	"github.com/amp-labs/connectors/test/snapchatads"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testMembers(ctx)
	if err != nil {
		return 1
	}

	err = testRoles(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testMembers(ctx context.Context) error {
	conn := snapchatads.GetConnector(ctx)

	slog.Info("Deleting the members")

	deleteParams := common.DeleteParams{
		ObjectName: "members",
		RecordId:   "97c5f4fa-816c-40de-bf77-4e1aa0555d1e",
	}

	deleteRes, err := Delete(ctx, conn, deleteParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(deleteRes); err != nil {
		return err
	}

	return nil
}

func testRoles(ctx context.Context) error {
	conn := snapchatads.GetConnector(ctx)

	slog.Info("Deleting the roles")

	deleteParams := common.DeleteParams{
		ObjectName: "roles",
		RecordId:   "1104b401-7488-4a77-8a65-5f403aea0bb4",
	}

	deleteRes, err := Delete(ctx, conn, deleteParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(deleteRes); err != nil {
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
