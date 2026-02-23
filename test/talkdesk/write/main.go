package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/talkdesk"
	"github.com/amp-labs/connectors/test/talkdesk"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := talkdesk.NewConnector(ctx)

	err := testCreatingAttributes(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdatingAttributes(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingAttributes(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "attributes",
		RecordData: map[string]any{
			"name":                "Attribute 4",
			"active":              true,
			"proficiency":         "none",
			"default_proficiency": 100,
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testUpdatingAttributes(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "attributes",
		RecordId:   "6f7d49f0-e29a-45dd-9946-4523d04ad740",
		RecordData: map[string]any{
			"name": "ABCD",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
