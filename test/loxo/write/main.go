package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/loxo"
	"github.com/amp-labs/connectors/test/loxo"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testCompanies(ctx)
	if err != nil {
		return 1
	}

	err = testPeople(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testCompanies(ctx context.Context) error {
	conn := loxo.GetLoxoConnector(ctx)

	slog.Info("Creating the companies")

	writeParams := common.WriteParams{
		ObjectName: "companies",
		RecordData: map[string]any{
			"company[name]": "DivTech",
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
	conn := loxo.GetLoxoConnector(ctx)

	slog.Info("Creating the people")

	writeParams := common.WriteParams{
		ObjectName: "people",
		RecordData: map[string]any{
			"person[name]": "Kumar",
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

	slog.Info("Updating the people")

	updateParams := common.WriteParams{
		ObjectName: "people",
		RecordData: map[string]any{
			"person[email]": "kumar123@gmail.com",
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
