package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/talkdesk"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := talkdesk.NewConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{"contacts", "do-not-call-lists", "ring-groups", "campaigns"})
	if err != nil {
		return err
	}

	utils.DumpJSON(m, os.Stdout)

	return nil
}
