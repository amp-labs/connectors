package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used
	"os"
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

type MeetingsPayload struct {
	StartTime      string    `json:"startTime"`
	EndTime        string    `json:"endTime"`
	Title          string    `json:"title"`
	Invitees       []Invitee `json:"invitees"`
	ExternalId     string    `json:"externalId"`
	OrganizerEmail string    `json:"organizerEmail"`
}

type Invitee struct {
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
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

	res := createMeetings(ctx, conn, &MeetingsPayload{
		StartTime:      "2025-11-17T02:30:00-08:00",
		EndTime:        "2025-11-17T03:30:00-08:00",
		Title:          "Created from Script",
		ExternalId:     createUniqueID(),
		OrganizerEmail: "integration.user+gong1@withampersand.com",
		Invitees: []Invitee{
			{
				DisplayName: "Test User",
				Email:       "test@test.com",
			},
		},
	})

	jsonStr, _ := json.MarshalIndent(res, "", "  ")

	slog.Info("Creating Meeting...")

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	res = createDigitalInteraction(ctx, conn, map[string]any{
		"eventId":   createUniqueID(),
		"timestamp": "2025-10-17T10:30:00.000Z",
		"eventType": "page viewed",
		"device":    "MOBILE",
		"content": map[string]any{
			"contentId":    createUniqueID(),
			"contentTitle": "Test Content from Script",
			"contentLabel": []string{"test", "script"},
			"contentUrl":   "https://example.com/test-content",
		},
		"trackingId": createUniqueID(),
	})

	jsonStr, _ = json.MarshalIndent(res, "", "  ")

	slog.Info("Creating Digital Interaction...")

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	slog.Info("Successful test completion")
}

func createCalls(ctx context.Context, conn *gong.Connector, payload *CallsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
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

func createMeetings(ctx context.Context, conn *gong.Connector, payload *MeetingsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "meetings",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Gong", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a Meeting")
	}

	return res
}

func createDigitalInteraction(ctx context.Context, conn *gong.Connector, payload any) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "digital-interaction",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Gong", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a Digital Interaction")
	}

	return res
}
