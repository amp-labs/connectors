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
	workspaceMemberID = "073f4c74-b60d-4de9-992a-0f799b5e442e"
	recordId          = "2db97cee-6c6b-4486-ae52-db8e4b6f44e9"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testLists(ctx)
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

	err = testCompanies(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testLists(ctx context.Context) error {
	conn := attio.GetAttioConnector(ctx)

	slog.Info("Creating the list")

	writeParams := common.WriteParams{
		ObjectName: "lists",
		RecordData: map[string]any{
			"data": map[string]interface{}{
				"workspace_access": "full-access",
				"name":             "Recruiting",
				"api_slug":         "recruiting",
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
			"data": map[string]interface{}{
				"name": "Recruit",
			},
		},
		RecordId: writeRes.Data["id"].(map[string]interface{})["list_id"].(string),
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
			"data": map[string]interface{}{
				"deadline_at": "2023-10-03T14:00:00.000000000Z",
			},
		},
		RecordId: writeRes.Data["id"].(map[string]interface{})["task_id"].(string),
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

func testCompanies(ctx context.Context) error {
	conn := attio.GetAttioConnector(ctx)

	slog.Info("Creating the record for Companies")

	writeParams := common.WriteParams{
		ObjectName: "companies",
		RecordData: map[string]any{
			"data": map[string]any{
				"values": map[string]any{
					"name":        "FireFox",
					"domains":     []string{"firefox.com"},
					"description": "Firefox is a free, open-source web browser developed by the Mozilla Corporation. It's known for its speed, privacy features, and customizable options, competing with browsers like Chrome and Safari.",
					"categories":  []string{"SAAS", "Web Services & Apps", "Internet"},
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

	slog.Info("Updating the record for Companies")

	updateParams := common.WriteParams{
		ObjectName: "companies",
		RecordData: map[string]any{
			"data": map[string]interface{}{
				"values": map[string]any{
					"categories": []string{"SAAS"},
				},
			},
		},
		RecordId: writeRes.Data["id"].(map[string]interface{})["record_id"].(string),
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
