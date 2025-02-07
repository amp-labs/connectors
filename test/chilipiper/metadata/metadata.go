package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/test/chilipiper"
)

func main() {
	ctx := context.Background()

	conn := chilipiper.GetChiliPiperConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"workspace", "team", "workspace_users", "meme"})
	if err != nil {
		slog.Error(err.Error())
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)
}
