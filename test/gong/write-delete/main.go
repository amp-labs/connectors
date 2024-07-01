package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/gong"
	connTest "github.com/amp-labs/connectors/test/gong"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

type callsPayload struct { // TODO fill in properties
	Name       string `json:"name,omitempty"`
	View       string `json:"view,omitempty"`
	ViewParams string `json:"view_params,omitempty"` // JSON object of list view parameters
	IsDefault  *bool  `json:"is_default,omitempty"`
	Shared     bool   `json:"shared,omitempty"`
}

var (
	objectName = "calls" // nolint: gochecknoglobals
)


func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("GONG_CRED_FILE")
	if filePath == "" {
		filePath = "./gong-creds.json"
	}

	conn := connTest.GetGongConnector(ctx, filePath)
	defer utils.Close(conn)

	fmt.Println("> TEST Create/Update/Delete calls")
	fmt.Println("Creating calls")

	// NOTE: list view must have unique `Name`
	view := createCalls(ctx, conn, &callsPayload{
		Name:       "Tom's Prospects",
		View:       "companies",
		ViewParams: "",
		IsDefault:  mockutils.Pointers.Bool(true),
		Shared:     false,
	})

	fmt.Println("Updating some calls properties")
	updateCalls(ctx, conn, view.RecordId, &callsPayload{
		Name:      "Jerry's Prospects",
		View:      "companies",
		IsDefault: mockutils.Pointers.Bool(false),
	})

	fmt.Println("View that calls has changed accordingly")

	res := readCalls(ctx, conn)

	updatedView := searchCalls(res, "id", view.RecordId)
	for k, v := range map[string]string{
		"name":       "Jerry's Prospects",
		"view":       "companies",
		"is_default": "false",
		"shared":     "false",
	} {
		if !mockutils.DoesObjectCorrespondToString(updatedView[k], v) {
			utils.Fail("error updated properties do not match", k, v, updatedView[k])
		}
	}

	fmt.Println("Removing this calls")
	removeCalls(ctx, conn, view.RecordId)
	fmt.Println("> Successful test completion")
}

func searchCalls(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Raw
		}
	}

	utils.Fail("error finding calls")

	return nil
}

func readCalls(ctx context.Context, conn *gong.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: []string{
			"id", "view", "name",
		},
	})
	if err != nil {
		utils.Fail("error reading from Gong", "error", err)
	}

	return res
}

func createCalls(ctx context.Context, conn *gong.Connector, payload *callsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Gong", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a calls")
	}

	return res
}

func updateCalls(ctx context.Context, conn *gong.Connector, viewID string, payload *callsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   viewID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Gong", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a calls")
	}

	return res
}

func removeCalls(ctx context.Context, conn *gong.Connector, viewID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   viewID,
	})
	if err != nil {
		utils.Fail("error deleting for Gong", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a calls")
	}
}
