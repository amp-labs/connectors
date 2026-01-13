package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	rc "github.com/amp-labs/connectors/providers/ringcentral"
	"github.com/amp-labs/connectors/test/ringcentral"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn, err := ringcentral.NewConnector(ctx)
	if err != nil {
		utils.Fail("error creating ringcentral connector", "error", err)
	}

	if err := testCreatingContact(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := testUpdateContact(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := testCreatingBridges(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func testCreatingContact(ctx context.Context, conn *rc.Connector) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"firstName":     "Charlie",
			"lastName":      "Williams",
			"businessPhone": "+15551234567",
			"businessAddress": map[string]any{
				"street": "20 Davis Dr.",
				"city":   "Belmont",
				"state":  "CA",
				"zip":    94002,
			},
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testUpdateContact(ctx context.Context, conn *rc.Connector) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordId:   "517566053",
		RecordData: map[string]any{
			"firstName": "Charles",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testCreatingBridges(ctx context.Context, conn *rc.Connector) error {
	params := common.WriteParams{
		ObjectName: "bridges",
		RecordData: map[string]any{
			"name":     "Monthly Meeting with Joseph",
			"type":     "Instant",
			"security": map[string]any{"passwordProtected": true, "password": "Wq123ygs15", "noGuests": false, "sameAccount": false, "e2ee": false},
			"preferences": map[string]any{"join": map[string]any{"audioMuted": false, "videoMuted": false, "waitingRoomRequired": "Nobody",
				"pstn": map[string]any{"promptAnnouncement": true, "promptParticipants": true}},
				"playTones":   "Off",
				"musicOnHold": true, "joinBeforeHost": true, "screenSharing": true, "recordingsMode": "User", "transcriptionsMode": "User",
				"recordings": map[string]any{"everyoneCanControl": map[string]any{"enabled": true, "locked": false}, "autoShared": map[string]any{"enabled": true, "locked": false}}, "allowEveryoneTranscribeMeetings": true},
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
