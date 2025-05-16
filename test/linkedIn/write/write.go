package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/linkedIn"
	"github.com/amp-labs/connectors/test/linkedIn"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testAdAccounts(ctx)
	if err != nil {
		return 1
	}

	err = TestAdTargetTemplates(ctx)
	if err != nil {
		return 1
	}

	err = testConversationAds(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testAdAccounts(ctx context.Context) error {
	conn := linkedIn.GetConnector(ctx)

	slog.Info("Creating the Ad Account")

	writeParams := common.WriteParams{
		ObjectName: "adAccounts",
		RecordData: map[string]any{
			"currency":                       "USD",
			"name":                           "Demo Account",
			"notifiedOnCampaignOptimization": true,
			"notifiedOnCreativeApproval":     true,
			"notifiedOnCreativeRejection":    true,
			"notifiedOnEndOfCampaign":        true,
			"reference":                      "urn:li:organization:2414183",
			"type":                           "BUSINESS",
			"test":                           true,
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

	slog.Info("updating the Ad Account")

	updateParams := common.WriteParams{
		ObjectName: "adAccounts",
		RecordData: map[string]any{
			"patch": map[string]any{
				"$set": map[string]any{
					"name": "This is a new account name.",
				},
			},
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

func TestAdTargetTemplates(ctx context.Context) error {
	conn := linkedIn.GetConnector(ctx)

	slog.Info("Creating the Ad target templates")

	writeParams := common.WriteParams{
		ObjectName: "adTargetTemplates",
		RecordData: map[string]any{
			"name":        "AI Audience Template",
			"description": "Tech Audience interested in Artificial Intelligence in North America",
			"account":     "urn:li:sponsoredAccount:514674276",
			"targetingCriteria": map[string]any{
				"include": map[string]any{
					"and": []map[string]any{
						{
							"or": map[string]any{
								"urn:li:adTargetingFacet:interests": []string{
									"urn:li:interest:308",
								},
							},
						},
					},
				},
				"exclude": map[string]any{
					"or": map[string]any{
						"urn:li:adTargetingFacet:seniorities": []string{
							"urn:li:seniority:1",
							"urn:li:seniority:2",
						},
					},
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

	slog.Info("Updating the Ad target templates")

	updateParams := common.WriteParams{
		ObjectName: "adTargetTemplates",
		RecordData: map[string]any{
			"patch": map[string]any{
				"$set": map[string]any{
					"name": "This is a new account name.",
				},
			},
		},
		RecordId: writeParams.RecordId,
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

func testConversationAds(ctx context.Context) error {
	conn := linkedIn.GetConnector(ctx)

	slog.Info("Creating conversation Ads")

	writeParams := common.WriteParams{
		ObjectName: "conversationAds",
		RecordData: map[string]any{
			"parentAccount": "urn:li:sponsoredAccount:514674276",
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
