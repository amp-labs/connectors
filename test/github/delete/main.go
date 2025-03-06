package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/github"
	hs "github.com/amp-labs/connectors/test/github"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := hs.GetGithubConnector(ctx)

	slog.Info("> TEST Delete gists")
	if err := deleteGist(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

}

func deleteGist(ctx context.Context, conn *github.Connector) error {
	config := common.DeleteParams{
		ObjectName: "gists",
		RecordId:   "d4e45740e4c6edd1a329436244718fc2", // Replace with a valid Gist ID
	}

	result, err := conn.Delete(ctx, config)
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
