package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	m "github.com/amp-labs/connectors/test/monday"
)

func main() {
	ctx := context.Background()

	// This will get all boards
	err := testReadBoards(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testReadBoardsPagination(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	// This will get all users
	err = testReadUsers(ctx)
	if err != nil {
		slog.Error(err.Error())
	}
}

func testReadBoards(ctx context.Context) error {
	conn := m.GetMondayConnector(ctx)

	params := common.ReadParams{
		ObjectName: "boards",
		Fields:     connectors.Fields("id", "name"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
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

func testReadBoardsPagination(ctx context.Context) error {
	conn := m.GetMondayConnector(ctx)

	params := common.ReadParams{
		ObjectName: "boards",
		Fields:     connectors.Fields("id", "name"),
		NextPage:   "1",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
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

func testReadUsers(ctx context.Context) error {
	conn := m.GetMondayConnector(ctx)

	params := common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("email", "id", "name"),
		Since:      time.Now().Add(-1800 * time.Hour),
		NextPage:   "1",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
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
