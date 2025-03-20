package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/aha"
	hs "github.com/amp-labs/connectors/test/aha"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := hs.GetAhaConnector(ctx)

	slog.Info("> TEST Creating Historic Audits")

	if err := createHistoricAudits(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("> TEST Creating Idea Users")

	recordId, err := createIdeaUsers(ctx, conn)

	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info("> TEST Updating Idea Users")

	if err := updateIdeaUsers(ctx, conn, recordId); err != nil {
		slog.Error(err.Error())
	}

}

func createHistoricAudits(ctx context.Context, conn *aha.Connector) error {
	config := common.WriteParams{
		ObjectName: "historical_audits",
		RecordData: map[string]any{

			"search": map[string]any{
				"created_since":  "2019-01-01T00:00:00Z",
				"created_before": "2019-01-01T00:00:00Z",
				"auditable_type": "Feature",
				"auditable_id":   1007868956,
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createIdeaUsers(ctx context.Context, conn *aha.Connector) (string, error) {
	config := common.WriteParams{
		ObjectName: "idea_users",
		RecordData: map[string]any{
			"idea_user": map[string]any{
				"email":      "samsdsfafsd@example.com",
				"first_name": "sam",
				"last_name":  "doe",
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	fmt.Println(string(jsonStr))

	return result.RecordId, nil
}

func updateIdeaUsers(ctx context.Context, conn *aha.Connector, recordId string) error {
	config := common.WriteParams{
		ObjectName: "idea_users",
		RecordId:   recordId,
		RecordData: map[string]any{
			"idea_user": map[string]any{
				"first_name": "sam" + time.Now().String(),
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}
