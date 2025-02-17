package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zoom"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoom"
)

var objectName = "users"

type CreateUserPaylod struct {
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

	conn := connTest.GetZoomConnector(ctx, zoom.ModuleUser)

	slog.Info("> TEST Create/Update/Delete User")
	slog.Info("Creating a user...")

	createUser(ctx, conn, &CreateUserPaylod{
		Action: "create",
		UserInfo: UserInfo{
			Email:       "tiger@gmail.com",
			FirstName:   "Tiger",
			LastName:    "Woods",
			DisplayName: "Tiger Woods",
			Password:    "password",
			Type:        1,
		},
	})

	slog.Info("> Successful test completion")
}

func createUser(ctx context.Context, conn *zoom.Connector, payload *CreateUserPaylod) {
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
}
