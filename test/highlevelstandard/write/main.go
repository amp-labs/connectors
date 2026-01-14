package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/highlevelstandard"
	"github.com/amp-labs/connectors/test/highlevelstandard"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testBusinesses(ctx)
	if err != nil {
		return 1
	}

	err = testProductsCollections(ctx)
	if err != nil {
		return 1
	}

	err = testCalendarsGroups(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testBusinesses(ctx context.Context) error {
	conn := highlevelstandard.GetHighLevelStandardConnector(ctx)

	slog.Info("Creating the Businesses")

	writeParams := common.WriteParams{
		ObjectName: "businesses",
		RecordData: map[string]any{
			"name":        "Demo",
			"locationId":  "iV1BEzddaWWLqU2kXhcN",
			"phone":       "+18832327657",
			"email":       "johnmalt@gmail.com",
			"website":     "www.xyz.com",
			"address":     "street address",
			"city":        "new york",
			"postalCode":  "12312312",
			"state":       "new york",
			"country":     "us",
			"description": "business description",
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

	slog.Info("Updating the Businesses")

	updateParams := common.WriteParams{
		ObjectName: "businesses",
		RecordData: map[string]any{
			"name": "Google",
		},
		RecordId: writeRes.RecordId,
	}

	updateRes, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(updateRes); err != nil {
		return err
	}

	return nil
}

func testProductsCollections(ctx context.Context) error {
	conn := highlevelstandard.GetHighLevelStandardConnector(ctx)

	slog.Info("Creating products collections")

	writeParams := common.WriteParams{
		ObjectName: "products/collections",
		RecordData: map[string]any{
			"altId":   "iV1BEzddaWWLqU2kXhcN",
			"altType": "location",
			"name":    "Product sellers",
			"slug":    "product-sellers",
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

func testCalendarsGroups(ctx context.Context) error {
	conn := highlevelstandard.GetHighLevelStandardConnector(ctx)

	slog.Info("Creating calendars groups")

	writeParams := common.WriteParams{
		ObjectName: "calendars/groups",
		RecordData: map[string]any{
			"locationId":  "iV1BEzddaWWLqU2kXhcN",
			"name":        "group b",
			"description": "group description",
			"slug":        "15-mins",
			"isActive":    true,
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

	slog.Info("Updating calendars groups")

	updateParams := common.WriteParams{
		ObjectName: "calendars/groups",
		RecordData: map[string]any{
			"name":        "group b",
			"description": "group description",
			"slug":        "30-mins",
		},
		RecordId: writeRes.RecordId,
	}

	updateRes, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(updateRes); err != nil {
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
