package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/chilipiper"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := chilipiper.GetChiliPiperConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"workspace", "team", "workspace/users", "meme"})
	if err != nil {
		slog.Error(err.Error())
	}

	// Print the results
	utils.DumpJSON(m, os.Stdout)
}
