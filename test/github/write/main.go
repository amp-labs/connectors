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

	slog.Info("> TEST Add User Emails")

	if err := addUserEmails(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("> TEST Create gists")

	if err := createGist(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("> TEST update user codespaces")

	if err := updateUserCodespaces(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func addUserEmails(ctx context.Context, conn *github.Connector) error {
	config := common.WriteParams{
		ObjectName: "user/emails",
		RecordData: map[string]any{
			"emails": []string{
				"testagain@withmampersand.com",
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createGist(ctx context.Context, conn *github.Connector) error {
	config := common.WriteParams{
		ObjectName: "gists",
		RecordData: map[string]any{
			"description": "Example of a gist",
			"public":      false,
			"files": map[string]any{
				"README.md": map[string]string{"content": "Hello World"},
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func updateUserCodespaces(ctx context.Context, conn *github.Connector) error {
	config := common.WriteParams{
		ObjectName: "user/codespaces",
		RecordId:   "fuzzy-fishstick-jj457pqw4q7g2565v",
		RecordData: map[string]any{
			"display_name": "test",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}
