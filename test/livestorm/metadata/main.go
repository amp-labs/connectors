package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/livestorm"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := connTest.GetLivestormConnector(ctx)

	meta, err := conn.ListObjectMetadata(ctx, []string{
		"events",
		"people",
		"people_attributes",
		"jobs",
		"session_chat_messages",
	})
	if err != nil {
		log.Fatalf("ListObjectMetadata error: %v", err)
	}

	utils.DumpJSON(meta, os.Stdout)
}
