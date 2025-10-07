package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/meta"
	connTest "github.com/amp-labs/connectors/test/meta"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testAdLabels(ctx)
	if err != nil {
		return 1
	}

	err = testBusinessUsers(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testAdLabels(ctx context.Context) error {
	conn := connTest.GetFacebookConnector(ctx)

	slog.Info("Creating the ad labels")

	writeParams := common.WriteParams{
		ObjectName: "adlabels",
		RecordData: map[string]any{
			"name": "Entertainment",
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

func testBusinessUsers(ctx context.Context) error {
	conn := connTest.GetFacebookConnector(ctx)

	slog.Info("Creating the business users")

	writeParams := common.WriteParams{
		ObjectName: "business_users",
		RecordData: map[string]any{
			"email": "sample@gmail.com",
			"role":  "EMPLOYEE",
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
