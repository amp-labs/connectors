package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/fireflies"
	"github.com/amp-labs/connectors/test/fireflies"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testAddToLiveMeeting(ctx)
	if err != nil {
		return 1
	}

	err = testCreateBite(ctx)
	if err != nil {
		return 1
	}

	err = testSetUserRole(ctx)
	if err != nil {
		return 1
	}

	err = testUploadAudio(ctx)
	if err != nil {
		return 1
	}

	err = testUpdateMeetingTitle(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testAddToLiveMeeting(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Add Live meeting")

	writeParams := common.WriteParams{
		ObjectName: "liveMeetings",
		RecordData: map[string]any{
			"meeting_link": "https://meet.google.com/qdt-vccw-nzt",
			// "title":            "demo",
			// "meeting_password": "Ab34TRD",
			// "duration":         60,
			// "language":         "en",
			// "attendees": []any{
			// 	map[string]string{
			// 		"displayName": "Fireflies Notetaker",
			// 		"email":       "notetaker@fireflies.ai",
			// 		"phoneNumber": "5522668874",
			// 	},
			// },
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	return nil
}

func testCreateBite(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Creating the bite")

	writeParams := common.WriteParams{
		ObjectName: "bites",
		RecordData: map[string]any{
			"transcriptId": "01JSXJ9T9DCS3PH46ACCRSCAX2",
			"startTime":    float64(3),
			"endTime":      float64(4),
			// "name":         "bite",
			// "media_type":   "audio",
			// "privacies":    []string{"team", "participants"},
			// "summary":      "creating the bites",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	return nil
}

func testSetUserRole(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Set user role")

	writeParams := common.WriteParams{
		ObjectName: "userRole",
		RecordData: map[string]any{
			"user_id": "01JSH43RZP1W6GAWQ2B87EAK7X",
			"role":    "user",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	return nil
}

func testUploadAudio(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Upload the audio file")

	var attendees = []any{
		map[string]string{
			"displayName": "Fireflies Notetaker",
			"email":       "notetaker@fireflies.ai",
			"phoneNumber": "5522668874",
		},
		map[string]string{
			"displayName": "Notetaker",
			"email":       "notetaker@fireflies.ai",
			"phoneNumber": "5246233652",
		},
	}

	inputParts := map[string]any{
		"url":        "https://www.nch.com.au/scribe/practice/audio-sample-4.mp3",
		"title":      "Medical Report",
		"attendees":  attendees,
		"save_video": true,
	}

	writeParams := common.WriteParams{
		ObjectName: "audio",
		RecordData: inputParts,
		RecordId:   "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	return nil
}

func testUpdateMeetingTitle(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Updating the meeting title")

	inputParts := map[string]any{
		"id":    "01JW6CPYTHM5DEFKH9X739BDPS",
		"title": "Daily Standup",
	}
	writeParams := common.WriteParams{
		ObjectName: "meetingTitle",
		RecordData: inputParts,
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	if err := constructResponse(writeRes); err != nil {
		return err
	}

	return nil
}

func Write(ctx context.Context, conn *ap.Connector, payload common.WriteParams) (*common.WriteResult, error) {
	res, err := conn.Write(ctx, payload)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// unmarshal the write response.
func constructResponse(res *common.WriteResult) error {
	jsonStr, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
