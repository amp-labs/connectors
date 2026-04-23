package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/bentley"
	conn "github.com/amp-labs/connectors/test/bentley"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	bentleyConn := conn.GetBentleyConnector(ctx)

	recordId, err := createITwins(ctx, bentleyConn)

	if err != nil {
		slog.Error(err.Error())
	}

	err = updateITwins(ctx, bentleyConn, recordId)
	if err != nil {
		slog.Error(err.Error())
	}

	recordId, err = createImodels(ctx, bentleyConn)
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Done")

}

func createITwins(ctx context.Context, conn *bentley.Connector) (string, error) {
	config := common.WriteParams{
		ObjectName: "itwins",
		RecordData: map[string]any{
			"class":              "Endeavor",
			"subClass":           "Project",
			"type":               "Construction Project",
			"number":             gofakeit.Numerify("###-###"),
			"displayName":        gofakeit.Sentence(3),
			"geographicLocation": "Exton, PA",
			"latitude":           40.028,
			"longitude":          -75.621,
			"ianaTimeZone":       "America/New_York",
			"dataCenterLocation": "East US",
			"status":             "Active",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	fmt.Println(string(jsonStr))
	return result.RecordId, nil
}

func updateITwins(ctx context.Context, conn *bentley.Connector, recordId string) error {
	config := common.WriteParams{
		ObjectName: "itwins",
		RecordId:   recordId,
		RecordData: map[string]any{
			"description": "This is an updated test ITwin",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))
	return err
}

func createImodels(ctx context.Context, conn *bentley.Connector) (string, error) {
	config := common.WriteParams{
		ObjectName: "imodels",
		RecordData: map[string]any{
			"iTwinId":      "3fc186ec-73a9-4a65-a55b-cffa4591dbdf",
			"name":         "Sun City Renewable-energy Plant",
			"description":  "Overall model of wind and solar farms in Sun City",
			"creationMode": "empty",
			"extent": map[string]any{
				"southWest": map[string]any{
					"latitude":  46.13267702834806,
					"longitude": 7.672120009938448,
				},
				"northEast": map[string]any{
					"latitude":  46.302763954781234,
					"longitude": 7.835541640797823,
				},
			},
			"template":     nil,
			"baselineFile": nil,
			"geographicCoordinateSystem": map[string]any{
				"horizontalCRSId": "EPSG:4326",
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	fmt.Println(string(jsonStr))
	return result.RecordId, nil
}
