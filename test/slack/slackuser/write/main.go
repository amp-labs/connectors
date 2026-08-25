package main

import (
	"context"
	"fmt"
	"os"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	slackshared "github.com/amp-labs/connectors/test/slack"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx := context.Background()

	conn := slackshared.NewConnector(ctx, providers.SlackUserScope)

	externalId := fmt.Sprintf("ext-id-%d", os.Getpid())
	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"calls",
		map[string]any{
			"join_url":           "https://example.com/join",
			"external_unique_id": externalId,
		},
		map[string]any{
			"title": "Updated Call Name",
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("join_url", "external_unique_id", "title"),
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"join_url":           "https://example.com/join",
				"external_unique_id": externalId,
				"title":              "Updated Call Name",
			},
		})
}
