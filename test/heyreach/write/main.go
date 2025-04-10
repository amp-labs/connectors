package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/heyreach"
	"github.com/amp-labs/connectors/test/heyreach"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testList(ctx)
	if err != nil {
		return 1
	}

	err = testAddLeadToCampaign(ctx)
	if err != nil {
		return 1
	}

	err = testAddLeadToList(ctx)
	if err != nil {
		return 1
	}

	err = testSendMessage(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testList(ctx context.Context) error {
	conn := heyreach.GetHeyreachConnector(ctx)

	slog.Info("Creating the empty list")

	writeParams := common.WriteParams{
		ObjectName: "list/CreateEmptyList",
		RecordData: map[string]any{
			"name": "My List",
			"type": "USER_LIST",
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

func testAddLeadToCampaign(ctx context.Context) error {
	conn := heyreach.GetHeyreachConnector(ctx)

	slog.Info("Add lead to existing campaign")

	writeParams := common.WriteParams{
		ObjectName: "campaign/AddLeadsToCampaignV2",
		RecordData: map[string]any{
			"campaignId": 120469,
			"accountLeadPairs": []any{
				map[string]any{
					"lead": map[string]any{
						"firstName":  "Hari",
						"lastName":   "Dinesh",
						"profileUrl": "https://www.linkedin.com/in/ACoAADdu6TsBifdNTeNoyZBIm__NVAE38FGapc4",
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

	return nil
}

func testAddLeadToList(ctx context.Context) error {
	conn := heyreach.GetHeyreachConnector(ctx)

	slog.Info("Add lead to existing list")

	writeParams := common.WriteParams{
		ObjectName: "list/AddLeadsToListV2",
		RecordData: map[string]any{
			"listId": 196432,
			"leadS": []any{
				map[string]any{
					"firstName":  "Hari",
					"lastName":   "Dinesh",
					"profileUrl": "https://www.linkedin.com/in/ACoAADdu6TsBifdNTeNoyZBIm__NVAE38FGapc4",
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

func testSendMessage(ctx context.Context) error {
	conn := heyreach.GetHeyreachConnector(ctx)

	slog.Info("Sending message to LinkedIn conversation")

	writeParams := common.WriteParams{
		ObjectName: "inbox/SendMessage",
		RecordData: map[string]any{
			"message":           "Hi",
			"subject":           "Info",
			"conversationId":    "2-YTQ4NjFkZWUtZTY1NS00ZWZhLTk0YTctYWVkMGNjYjFlYTBkXzEwMA==",
			"linkedInAccountId": 72866,
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
