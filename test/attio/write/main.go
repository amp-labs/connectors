package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/test/attio"
)

const (
	workspaceMemberID = "67af46e4-a450-4fee-a1d1-39729b3af771"
	recordId          = "ec902ed9-aab7-4347-8e26-dca240ffba08"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testObjects(ctx)
	if err != nil {
		return 1
	}

	err = testLists(ctx)
	if err != nil {
		return 1
	}

	err = testNotes(ctx)
	if err != nil {
		return 1
	}

	err = testTasks(ctx)
	if err != nil {
		return 1
	}

	err = testWebhooks(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testObjects(ctx context.Context) error {
	conn := attio.GetAttioConnector(ctx)

	slog.Info("Creating the object")

	params := common.WriteParams{
		ObjectName: "objects",
		RecordData: map[string]any{
			"data": map[string]string{
				"api_slug":      "deal",
				"singular_noun": "Deals",
				"plural_noun":   "Dealss",
			},
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, params)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	slog.Info("Updating the object")

	updateparams := common.WriteParams{
		ObjectName: "objects",
		RecordData: map[string]any{
			"data": map[string]string{
				"singular_noun": "Deal",
			},
		},
		RecordId: writeRes.Data["id"].(map[string]any)["object_id"].(string),
	}

	updateres, err := Write(ctx, conn, updateparams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(updateres); err != nil {
		return err
	}

	return nil
}

func testLists(ctx context.Context) error {
	conn := attio.GetAttioConnector(ctx)

	slog.Info("Creating the list")

	writeParams := common.WriteParams{
		ObjectName: "lists",
		RecordData: map[string]any{
			"data": map[string]any{
				"workspace_access": "full-access",
				"name":             "Marketing",
				"api_slug":         "marketing_1",
				"parent_object":    "companies",
				"workspace_member_access": []map[string]string{
					{
						"workspace_member_id": workspaceMemberID,
						"level":               "full-access",
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

	slog.Info("Updating the list")

	updateParams := common.WriteParams{
		ObjectName: "lists",
		RecordData: map[string]any{
			"data": map[string]any{
				"name": "Sales",
			},
		},
		RecordId: writeRes.Data["id"].(map[string]any)["list_id"].(string),
	}

	writeres, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeres); err != nil {
		return err
	}

	return nil
}

func testNotes(ctx context.Context) error {
	conn := attio.GetAttioConnector(ctx)

	slog.Info("Creating the notes")

	writeParams := common.WriteParams{
		ObjectName: "notes",
		RecordData: map[string]any{
			"data": map[string]string{
				"format":           "plaintext",
				"parent_object":    "companies",
				"parent_record_id": recordId,
				"title":            "Call summary",
				"content":          "summary",
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

func testTasks(ctx context.Context) error {
	conn := attio.GetAttioConnector(ctx)

	slog.Info("Creating the task")

	writeParams := common.WriteParams{
		ObjectName: "tasks",
		RecordData: map[string]any{
			"data": map[string]any{
				"format":       "plaintext",
				"is_completed": false,
				"content":      "view summary",
				"deadline_at":  "2023-10-04T15:00:00.000000000Z",
				"linked_records": []map[string]any{
					{
						"target_object":    "companies",
						"target_record_id": recordId,
					},
				},
				"assignees": []map[string]any{
					{
						"referenced_actor_type": "workspace-member",
						"referenced_actor_id":   workspaceMemberID,
					},
					{
						"workspace_member_email_address": "integration.user@withampersand.com",
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

	slog.Info("Updating the tasks")

	updateParams := common.WriteParams{
		ObjectName: "tasks",
		RecordData: map[string]any{
			"data": map[string]any{
				"deadline_at": "2023-10-03T14:00:00.000000000Z",
			},
		},
		RecordId: writeRes.Data["id"].(map[string]any)["task_id"].(string),
	}

	writeres, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeres); err != nil {
		return err
	}

	return nil
}

func testWebhooks(ctx context.Context) error {
	conn := attio.GetAttioConnector(ctx)

	slog.Info("Creating the webhooks")

	writeParams := common.WriteParams{
		ObjectName: "webhooks",
		RecordData: map[string]any{
			"data": map[string]any{
				"target_url": "https://f87a-117-216-131-16.ngrok-free.app",
				"subscriptions": []map[string]any{
					{
						"event_type": "note.deleted",
						"filter":     nil,
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

	slog.Info("Updating the webhooks")

	updateParams := common.WriteParams{
		ObjectName: "webhooks",
		RecordData: map[string]any{
			"data": map[string]any{
				"subscriptions": []map[string]any{
					{
						"event_type": "note.created",
						"filter":     nil,
					},
				},
			},
		},
		RecordId: writeRes.Data["id"].(map[string]any)["webhook_id"].(string),
	}

	writeres, err := Write(ctx, conn, updateParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeres); err != nil {
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
