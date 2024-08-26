package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/atlassian"
	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
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
	issueTypeID = "10005"
)

// For this script replace project id and issue types with your values.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAtlassianConnector(ctx)
	defer utils.Close(conn)

	slog.Info("> TEST Create/Update/Delete issue")
	slog.Info("Creating issue")

	view := createIssue(ctx, conn, &issuePayload{
		Fields: issueFields{
			Project: identifier{
				Id: projectID,
			},
			Issuetype: identifier{
				Id: issueTypeID,
			},
			Summary: "The very new title of Jira issue",
		},
	})

	slog.Info("Updating some issue properties")

	newTitle := "Fix button dropdown in main menu"
	updateIssue(ctx, conn, view.RecordId, &issuePayload{
		Fields: issueFields{
			Project: identifier{
				Id: projectID,
			},
			Issuetype: identifier{
				Id: issueTypeID,
			},
			Summary: newTitle,
		},
	})

	slog.Info("View that issue has changed accordingly")

	res := readIssue(ctx, conn)

	updatedView := searchIssue(res, "id", view.RecordId)

	summaryProperty, ok := updatedView["summary"]
	if !ok {
		utils.Fail("couldn't find summary property")
	}

	if summaryProperty != newTitle {
		utils.Fail("error updated Jira Issue Title does not match")
	}

	slog.Info("Removing this issue")
	removeIssue(ctx, conn, view.RecordId)
	slog.Info("> Successful test completion")
}

func searchIssue(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Raw
		}
	}

	utils.Fail("error finding issue")

	return nil
}

func readIssue(ctx context.Context, conn *atlassian.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		Fields: []string{
			"id", "fields",
		},
	})
	if err != nil {
		utils.Fail("error reading from Atlassian", "error", err)
	}

	return res
}

func createIssue(ctx context.Context, conn *atlassian.Connector, payload *issuePayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Atlassian", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a issue")
	}

	return res
}

func updateIssue(ctx context.Context, conn *atlassian.Connector, viewID string, payload *issuePayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		RecordId:   viewID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Atlassian", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a issue")
	}

	return res
}

func removeIssue(ctx context.Context, conn *atlassian.Connector, viewID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		RecordId: viewID,
	})
	if err != nil {
		utils.Fail("error deleting for Atlassian", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a issue")
	}
}
