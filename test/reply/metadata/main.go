package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/reply"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := reply.GetReplyConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{
		"contacts", "contact-accounts", "contact-lists", "contact-account-lists",
		"sequences", "sequence-folders", "tasks",
		"email-templates", "email-template-folders", "email-accounts",
		"custom-fields", "schedules", "linkedin-accounts", "inbox/threads",
	})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
