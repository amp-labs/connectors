package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/blueshift"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := blueshift.GetBlueshiftConnector(ctx)

	slog.Info("Reading campaigns")

	res, err := connector.Read(ctx, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     datautils.NewStringSet("uuid", "name"),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading email templates")

	res, err = connector.Read(ctx, common.ReadParams{
		ObjectName: "email_templates",
		Fields:     datautils.NewStringSet("uuid", "name"),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading sms templates")

	res, err = connector.Read(ctx, common.ReadParams{
		ObjectName: "sms_templates",
		Fields:     datautils.NewStringSet("uuid", "author"),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading onsite slots")

	res, err = connector.Read(ctx, common.ReadParams{
		ObjectName: "onsite_slots",
		Fields:     datautils.NewStringSet("uuid", "name"),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
