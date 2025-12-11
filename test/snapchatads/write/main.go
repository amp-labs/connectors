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
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	conn := snapchatads.GetConnector(ctx)

	_, err := conn.GetPostAuthInfo(ctx)
	if err != nil {
		utils.Fail(err.Error())
	}

	err = testBillingCenters(ctx, conn)
	if err != nil {
		return 1
	}

	err = testMembers(ctx, conn)
	if err != nil {
		return 1
	}

	return 0
}

func testBillingCenters(ctx context.Context, conn *ap.Connector) error {
	slog.Info("Creating the billing centers")

	writeParams := common.WriteParams{
		ObjectName: "billingcenters",
		RecordData: map[string]any{
			"billingcenters": []any{
				map[string]any{
					"organization_id":                 "5cf59a25-5063-40e1-826b-5ceaf369b207",
					"name":                            "Kianjous Billing Center",
					"email_address":                   "honeybear_ltd@example.com",
					"address_line_1":                  "11 Honey Bear Road",
					"locality":                        "London",
					"administrative_district_level_1": "GB-LND",
					"country":                         "GB",
					"postal_code":                     "NW1 4RY",
				},
			},
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

	slog.Info("Updating the billing centers")

	updateParams := common.WriteParams{
		ObjectName: "billingcenters",
		RecordData: map[string]any{
			"billingcenters": []any{
				map[string]any{
					"id":                              "6dfb86f9-7c5f-4348-aa2b-f17f070db0bb",
					"organization_id":                 "5cf59a25-5063-40e1-826b-5ceaf369b207",
					"name":                            "Kianjous Billing Center",
					"email_address":                   "honeybear_ltd@example.com",
					"address_line_1":                  "11 Rotrary Road",
					"locality":                        "London",
					"administrative_district_level_1": "GB-LND",
					"country":                         "GB",
					"postal_code":                     "NW1 4RY",
					"alternative_email_addresses": []any{
						"mr_duck@example.com",
					},
				},
			},
		},
		RecordId: "6dfb86f9-7c5f-4348-aa2b-f17f070db0bb",
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

func testMembers(ctx context.Context, conn *ap.Connector) error {
	slog.Info("Creating the members")

	writeParams := common.WriteParams{
		ObjectName: "members",
		RecordData: map[string]any{
			"members": []any{
				map[string]any{
					"email":           "honeybear@example.com",
					"organization_id": "5cf59a25-5063-40e1-826b-5ceaf369b207",
					"display_name":    "Honey Bear",
				},
			},
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
