package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cl "github.com/amp-labs/connectors/providers/callrail"
	"github.com/amp-labs/connectors/test/callrail"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := callrail.NewConnector(ctx)
	conn.GetPostAuthInfo(ctx)

	err := testCreatingTags(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdatingCompanies(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingCompanies(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingTags(ctx context.Context, conn *cl.Connector) error {
	params := common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]any{
			"name":       "Existing Customer 1",
			"company_id": "COM019b4fe752f4700dae00d75433481573",
			"color":      "gray1",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testUpdatingCompanies(ctx context.Context, conn *cl.Connector) error {
	params := common.WriteParams{
		ObjectName: "companies",
		RecordId:   "COM019b4feb89a07e0eba5a04dfc3edcfd4",
		RecordData: map[string]any{
			"callscribe_enabled": true,
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func testCreatingCompanies(ctx context.Context, conn *cl.Connector) error {
	params := common.WriteParams{
		ObjectName: "companies",
		RecordData: map[string]any{
			"name":      "Widget Shop",
			"time_zone": "America/New_York",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
