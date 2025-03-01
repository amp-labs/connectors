package main

import (
	"context"
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

	res, err := connector.Read(ctx, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     datautils.NewStringSet("uuid", "name"),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
