package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	devrev "github.com/amp-labs/connectors/providers/devrev"
	connTest "github.com/amp-labs/connectors/test/devrev"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := connTest.GetConnector(ctx)

	recordID, err := testCreatingArticle(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdatingArticle(ctx, conn, recordID)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingArticle(ctx context.Context, conn *devrev.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "articles",
		RecordData: map[string]any{
			"title":       "Test Article",
			"resource":    map[string]any{"url": "https://www.example.com"},
			"owned_by":    []string{"DEVU-120"},
			"description": "Test Article Description",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

func testUpdatingArticle(ctx context.Context, conn *devrev.Connector, recordID string) error {
	params := common.WriteParams{
		ObjectName: "articles",
		RecordId:   recordID,
		RecordData: map[string]any{
			"description": "Test Article Description Updated",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
