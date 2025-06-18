package main

import (
	"context"
	"os"

	"github.com/amp-labs/connectors/test/lever"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := lever.GetConnector(ctx)

	m, err := connector.ListObjectMetadata(ctx, []string{
		"feedback",
		"files",
		"interviews",
		"notes",
		"offers",
		"panels",
		"forms",
		"referrals",
		"resumes",
		"archive_reasons",
		"audit_events",
		"sources",
		"stages",
		"tags",
		"users",
		"feedback_templates",
		"opportunities",
		"postings",
		"form_templates",
		"requisitions",
		"requisition_fields",
	})
	if err != nil {
		utils.Fail(err.Error())
	}

	utils.DumpJSON(m, os.Stdout)
}
