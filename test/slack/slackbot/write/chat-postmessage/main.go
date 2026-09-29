package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	slackshared "github.com/amp-labs/connectors/test/slack"
	"github.com/amp-labs/connectors/test/utils"
)

// Posts a message to the first channel the bot is a member of.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := slackshared.NewConnector(ctx, providers.Slack)

	channels, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "conversations",
		Fields:     connectors.Fields("id", "is_member"),
	})
	if err != nil {
		utils.Fail("error reading conversations", "error", err)
	}

	var channelID string

	for _, row := range channels.Data {
		if row.Fields["is_member"] == true {
			channelID = row.Id

			break
		}
	}

	if channelID == "" {
		utils.Fail("the bot is not a member of any channel")
	}

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "chat.postMessage",
		RecordData: map[string]any{
			"channel": channelID,
			"text":    "Sent by the Slack connector write live test.",
		},
	})
	if err != nil {
		utils.Fail("error posting message", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
