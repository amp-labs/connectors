package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/amp-labs/connectors/common"
	mk "github.com/amp-labs/connectors/test/marketo"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	err := testReadChannels(context.Background())
	if err != nil {
		return 1
	}

	err = testReadSmartCampaigns(context.Background())
	if err != nil {
		return 1
	}

	return 0
}

func testReadChannels(ctx context.Context) error {
	conn := mk.GetMarketoConnector(ctx)

	params := common.ReadParams{
		ObjectName: "channels",
		Fields:     []string{"applicableProgramType", "id", "name"},
		Since:      time.Now().Add(-720 * time.Hour),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadSmartCampaigns(ctx context.Context) error {
	conn := mk.GetMarketoConnector(ctx)

	params := common.ReadParams{
		ObjectName: "smartCampaigns",
		Fields:     []string{"description", "id", "name"},
		Since:      time.Now().Add(-1020 * time.Hour),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
