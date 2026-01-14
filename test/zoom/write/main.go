package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zoom"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoom"
)

var objectName = "users"

type CreateUserPayload struct {
	Action   string   `json:"action"`
	UserInfo UserInfo `json:"user_info"`
}

type UserInfo struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
	Type        int    `json:"type"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZoomConnector(ctx)

	slog.Info("> TEST Create User")
	slog.Info("Creating a user...")

	createUser(ctx, conn, &CreateUserPayload{
		Action: "create",
		UserInfo: UserInfo{
			Email:       "peter@gmail.com",
			FirstName:   "peter",
			LastName:    "parker",
			DisplayName: "peter parker",
			Password:    "uncleben123",
			Type:        1,
		},
	})

	slog.Info("> Successful test completion")
}

func createUser(ctx context.Context, conn *zoom.Connector, payload *CreateUserPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to zoom", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a user")
	}

	utils.DumpJSON(res, os.Stdout)
}
