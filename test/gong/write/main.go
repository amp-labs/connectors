package main

import (
	"context"
	"log/slog"
	"math/rand"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/gong"
	connTest "github.com/amp-labs/connectors/test/gong"
	"github.com/amp-labs/connectors/test/utils"
)

type CallsPayload struct {
	ClientUniqueId string      `json:"clientUniqueId"`
	ActualStart    string      `json:"actualStart"`
	Title          string      `json:"title"`
	Direction      string      `json:"direction"`
	PrimaryUser    string      `json:"primaryUser"`
	Parties        []CallParty `json:"parties"`
}

type CallParty struct {
	EmailAddress string `json:"emailAddress,omitempty"`
	UserId       string `json:"userId,omitempty"`
}

var objectName = "calls" // nolint: gochecknoglobals

// This script creates the Call object.
// Gong takes some time to process the call before it can be viewed on the dashboard
// or retrieved via a GET API request. Deletion, in order to clean up, is only available on the dashboard.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGongConnector(ctx)
	defer utils.Close(conn)

	slog.Info("TEST Create Call")
	slog.Info("Creating Call")

	createCalls(ctx, conn, &CallsPayload{
		ClientUniqueId: createUniqueID(),
		ActualStart:    "2021-02-17T02:30:00-08:00",
		Title:          "Created from Script",
		Direction:      "Inbound",
		PrimaryUser:    "2860266319383544353",
		Parties: []CallParty{
			{
				EmailAddress: "test@test.com",
			},
			{
				UserId: "2860266319383544353",
			},
		},
	})

	slog.Info("Successful test completion")
}

func createCalls(ctx context.Context, conn *gong.Connector, payload *CallsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Gong", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a Call")
	}

	return res
}

func createUniqueID() string {
	minV := 1
	maxV := 10000
	uniqueID := strconv.Itoa(rand.Intn(maxV-minV+1) + minV)
	return uniqueID
}
