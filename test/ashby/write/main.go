package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/ashby"
	hs "github.com/amp-labs/connectors/test/ashby"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := hs.GetAshbyConnector(ctx)

	slog.Info("> TEST Create Candidate")

	recordId, err := createCandidate(ctx, conn)
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info("> TEST Create Application")

	if err := createApplication(ctx, conn, recordId); err != nil {
		slog.Error(err.Error())
	}
	slog.Info("Done")

}

func createApplication(ctx context.Context, conn *ashby.Connector, candidateId string) error {
	config := common.WriteParams{
		ObjectName: "application",
		RecordData: map[string]any{
			"candidateId": candidateId,
			"jobId":       "783338ea-e6ac-406a-853b-964fa75a5d62",
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

func createCandidate(ctx context.Context, conn *ashby.Connector) (string, error) {
	config := common.WriteParams{
		ObjectName: "candidate",
		RecordData: map[string]any{
			"name": "Deepu",
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
