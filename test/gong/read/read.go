package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/gong"
	"github.com/joho/godotenv"
)

const (
	DefaultCredsFile = "creds.json"
)

func main() {
	os.Exit(mainFn())
}

func mainFn() int {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file provided")
	}

	conn := connTest.GetGongConnector(context.Background(), DefaultCredsFile)

	config := connectors.ReadParams{
		ObjectName: "calls", // could be calls, users
		Fields:     []string{"url"},
	}

	result, err := conn.Read(context.Background(), config)
	if err != nil {
		slog.Error("Error reading from Gong", "error", err)
		return 1
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		slog.Error("Marshalling Error", "error", err)
		return 1
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return 0
}
