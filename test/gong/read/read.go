package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/amp-labs/connectors/gong"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/utils"
	"github.com/joho/godotenv"
)

const (
	DefaultCredsFile = "creds.json"
)

func GetGongConnector(ctx context.Context, filePath string) *gong.Connector {

	registry := utils.NewCredentialsRegistry()

	readers := []utils.Reader{
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['clientId']",
			CredKey:  "clientId",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['clientSecret']",
			CredKey:  "clientSecret",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['refreshToken']",
			CredKey:  "refreshToken",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['accessToken']",
			CredKey:  "accessToken",
		},
		&utils.JSONReader{
			FilePath: filePath,
			JSONPath: "$['provider']",
			CredKey:  "provider",
		},
	}
	registry.AddReaders(readers...)

	cfg := utils.GongOAuthConfigFromRegistry(registry)
	tok := utils.GongOauthTokenFromRegistry(registry)

	conn, err := connectors.Gong(
		gong.WithClient(ctx, http.DefaultClient, cfg, tok),
	)
	if err != nil {
		slog.Error("error creating gong connector", "error", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	return conn
}

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

	gong := GetGongConnector(context.Background(), DefaultCredsFile)

	config := connectors.ReadParams{
		ObjectName: "calls", // could be calls, users
		Fields:     []string{"url"},
	}

	result, err := gong.Read(context.Background(), config)
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
