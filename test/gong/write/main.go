package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"log/slog"
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

	slog.Info("TEST Create Call")
	slog.Info("Creating Call")

	id, err := uniqueId(1, 10000)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	createCalls(ctx, conn, &CallsPayload{
		ClientUniqueId: strconv.Itoa(id),
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

func uniqueId(x, y int) (int, error) {
	if x+1 >= y {
		return x + 1, nil
	}

	rangeSize := y - x

	// Read random bytes (8 bytes for 64-bit value)
	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return 0, err
	}

	randUint := binary.BigEndian.Uint64(buf[:])

	result := x + 1 + int(randUint%uint64(rangeSize))

	return result, nil
}
