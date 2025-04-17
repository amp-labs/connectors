package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/pinterest"
	"github.com/amp-labs/connectors/test/pinterest"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	conn := pinterest.GetConnector(context.Background())

	err := testReadPins(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadBoards(context.Background(), conn)
	if err != nil {
		return 1
	}

	err = testReadMedia(context.Background(), conn)
	if err != nil {
		return 1
	}

	return 0
}

func testReadPins(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "pins",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadBoards(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "boards",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadMedia(ctx context.Context, conn *ap.Connector) error {
	params := common.ReadParams{
		ObjectName: "media",
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
