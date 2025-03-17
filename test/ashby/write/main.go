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

	slog.Info("> TEST Create Application")

	if err := createApplication(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("> TEST Create Candidate Note")

	if err := createCandidateNote(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Done")
}

func createApplication(ctx context.Context, conn *ashby.Connector) error {
	config := common.WriteParams{
		ObjectName: "application.create",
		RecordData: map[string]any{
			"candidateId": "bd7229a1-be3e-4e30-a538-0b95a41602d7",
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

func createCandidateNote(ctx context.Context, conn *ashby.Connector) error {
	config := common.WriteParams{
		ObjectName: "candidate.createNote",
		RecordData: map[string]any{
			"candidateId":       "bd7229a1-be3e-4e30-a538-0b95a41602d7",
			"sendNotifications": false,
			"note":              "this is a not what else ",
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
