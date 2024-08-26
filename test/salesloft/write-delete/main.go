package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesloft"
	msTest "github.com/amp-labs/connectors/test/salesloft"
	"github.com/amp-labs/connectors/test/utils"
)

type ListViewPayload struct {
	Name       string `json:"name,omitempty"`
	View       string `json:"view,omitempty"`
	ViewParams string `json:"view_params,omitempty"` // JSON object of list view parameters
	IsDefault  *bool  `json:"is_default,omitempty"`
	Shared     bool   `json:"shared,omitempty"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("SALESLOFT_CRED_FILE")
	if filePath == "" {
		filePath = "./salesloft-creds.json"
	}

	conn := msTest.GetSalesloftConnector(ctx, filePath)
	defer utils.Close(conn)

	fmt.Println("> TEST Create/Update/Delete ListView")
	fmt.Println("Creating ListView")

	// NOTE: list view must have unique `Name`
	view := createListView(ctx, conn, &ListViewPayload{
		Name:       "Tom's Prospects",
		View:       "companies",
		ViewParams: "",
		IsDefault:  boolPtr(true),
		Shared:     false,
	})

	fmt.Println("Updating some ListView properties")
	updateListView(ctx, conn, view.RecordId, &ListViewPayload{
		Name:      "Jerry's Prospects",
		View:      "companies",
		IsDefault: boolPtr(false),
	})

	fmt.Println("View that ListView has changed accordingly")

	res := readListViews(ctx, conn)

	updatedView := searchListView(res, "id", view.RecordId)
	for k, v := range map[string]string{
		"name":       "Jerry's Prospects",
		"view":       "companies",
		"is_default": "false",
		"shared":     "false",
	} {
		if !compare(updatedView[k], v) {
			utils.Fail("error updated properties do not match", k, v, updatedView[k])
		}
	}

	fmt.Println("Removing this ListView")
	removeListView(ctx, conn, view.RecordId)
	fmt.Println("> Successful test completion")
}

func searchListView(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if compare(data.Fields[key], value) {
			return data.Raw
		}
	}

	utils.Fail("error finding ListView")

	return nil
}

func readListViews(ctx context.Context, conn *salesloft.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "saved_list_views",
		Fields: []string{
			"id", "view", "name",
		},
	})
	if err != nil {
		utils.Fail("error reading from Salesloft", "error", err)
	}

	return res
}

func createListView(ctx context.Context, conn *salesloft.Connector, payload *ListViewPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "saved_list_views",
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Salesloft", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a ListView")
	}

	return res
}

func updateListView(ctx context.Context, conn *salesloft.Connector, viewID string, payload *ListViewPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "saved_list_views",
		RecordId:   viewID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Salesloft", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a ListView")
	}

	return res
}

func removeListView(ctx context.Context, conn *salesloft.Connector, viewID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "saved_list_views",
		RecordId:   viewID,
	})
	if err != nil {
		utils.Fail("error deleting for Salesloft", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a ListView")
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func compare(field any, value string) bool {
	if len(value) == 0 && field == nil {
		return true
	}

	return fmt.Sprintf("%v", field) == value
}
