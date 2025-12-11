package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/fireflies"
	"github.com/amp-labs/connectors/test/fireflies"
	"github.com/amp-labs/connectors/test/utils"
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

	err = testUpdateMeetingPrivacy(ctx)
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
			"meeting_link":     "https://meet.google.com/hey-gdmi-xht",
			"title":            "demo",
			"meeting_password": "Ab34TRD",
			"duration":         60,
			"language":         "en",
			"attendees": []any{
				map[string]string{
					"displayName": "Fireflies Notetaker",
					"email":       "notetaker@fireflies.ai",
					"phoneNumber": "5522668874",
				},
			},
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	utils.DumpJSON(writeRes, os.Stdout)

	return nil
}

func testCreateBite(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Creating the bite")

	writeParams := common.WriteParams{
		ObjectName: "bites",
		RecordData: map[string]any{
			"transcriptId": "01K9YQ1SPTN3X9RAEXV1AH2P6G",
			"startTime":    float64(3),
			"endTime":      float64(4),
			"name":         "bite",
			"media_type":   "audio",
			"privacies":    []string{"team", "participants"},
			"summary":      "creating the bites",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	utils.DumpJSON(writeRes, os.Stdout)

	return nil
}

func testSetUserRole(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Set user role")

	writeParams := common.WriteParams{
		ObjectName: "userRole",
		RecordData: map[string]any{
			"user_id": "YUBzRk85N2",
			"role":    "user",
		},
		RecordId: "",
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	utils.DumpJSON(writeRes, os.Stdout)

	return nil
}

func testUploadAudio(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Upload the audio file")

	attendees := []any{
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

	utils.DumpJSON(writeRes, os.Stdout)

	return nil
}

func testUpdateMeetingTitle(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Updating the meeting title")

	inputParts := map[string]any{
		"id":    "01K9YQ1SPTN3X9RAEXV1AH2P6G",
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

	utils.DumpJSON(writeRes, os.Stdout)

	return nil
}

func testUpdateMeetingPrivacy(ctx context.Context) error {
	conn := fireflies.GetFirefliesConnector(ctx)

	slog.Info("Updating the meeting privacy")

	inputParts := map[string]any{
		"id":      "01K9YQ1SPTN3X9RAEXV1AH2P6G",
		"privacy": "owner",
	}
	writeParams := common.WriteParams{
		ObjectName: "meetingPrivacy",
		RecordData: inputParts,
	}

	writeRes, err := Write(ctx, conn, writeParams)
	if err != nil {
		fmt.Println("ERR: ", err)

		return err
	}

	utils.DumpJSON(writeRes, os.Stdout)

	return nil
}

func Write(ctx context.Context, conn *ap.Connector, payload common.WriteParams) (*common.WriteResult, error) {
	res, err := conn.Write(ctx, payload)
	if err != nil {
		return nil, err
	}

	return res, nil
}
