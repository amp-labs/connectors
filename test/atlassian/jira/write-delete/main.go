package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

type issuePayload struct {
	Fields issueFields `json:"fields"`
}

type issueFields struct {
	Project   identifier `json:"project"`
	Issuetype identifier `json:"issuetype"`
	Summary   string     `json:"summary"`
}

type identifier struct {
	Id string `json:"id"`
}

const (
	projectID   = "10001"
	issueTypeID = "10007"
)

// For this script replace project id and issue types with your values.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAtlassianConnector(ctx)

	oldTitle := "The very new title of Jira issue"
	newTitle := "Fix button dropdown in main menu"

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"issues",
		issuePayload{
			Fields: issueFields{
				Project: identifier{
					Id: projectID,
				},
				Issuetype: identifier{
					Id: issueTypeID,
				},
				Summary: oldTitle,
			},
		},
		issuePayload{
			Fields: issueFields{
				Project: identifier{
					Id: projectID,
				},
				Issuetype: identifier{
					Id: issueTypeID,
				},
				Summary: newTitle,
			},
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "summary"),
			SearchBy: testscenario.Property{
				Key:   "summary",
				Value: oldTitle,
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"summary": newTitle,
			},
		},
	)
}
