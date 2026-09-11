package main

import (
	"context"
	"log"
	"os"

	mailgun "github.com/amp-labs/connectors/test/mailgun"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := mailgun.GetMailgunConnector(ctx)

	objectNames := []string{
		// Domain-scoped (domain supplied via connector metadata).
		"bounces",
		"complaints",
		"unsubscribes",
		"whitelists",
		"tags",
		"domains/credentials",
		"templates",
		"domains/keys",

		// Account-scoped.
		"domains",
		"routes",
		"lists",
		"lists/members",
		"forwards",
		"ip_pools",
		"ips",
		"keys",
		"dkim/keys",
		"webhooks",
		"accounts/subaccounts",
		"accounts/subaccounts/ip_pools/all",
		"users",
		"ip_whitelist",
		"dynamic_pools/domains",
		"thresholds/limits",
		"thresholds/alerts/send",
		"account/templates",

		// POST-sourced.
		"analytics/logs",
	}

	m, err := connector.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
