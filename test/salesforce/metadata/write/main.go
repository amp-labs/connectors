package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/salesforce"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	conn := connTest.GetSalesforceConnector(ctx)

	input := []byte(`
<createMetadata>
    <metadata xsi:type="CustomObject">
        <fullName>TestObject15__c</fullName>
        <label>Test Object 15</label>
        <pluralLabel>Test Objects 15</pluralLabel>
        <nameField>
            <type>Text</type>
            <label>Test Object Name</label>
        </nameField>
        <deploymentStatus>Deployed</deploymentStatus>
        <sharingModel>ReadWrite</sharingModel>
    </metadata>
    <metadata xsi:type="CustomField">
        <fullName>TestObject13__c.Comments__c</fullName>
        <label>Comments</label>
        <type>LongTextArea</type>
        <length>500</length>
        <inlineHelpText>This field contains help text for this object</inlineHelpText>
        <description>Add your comments about this object here</description>
        <visibleLines>30</visibleLines>
        <required>false</required>
        <trackFeedHistory>false</trackFeedHistory>
        <trackHistory>false</trackHistory>
    </metadata>
</createMetadata>
`)
	accessToken := connTest.GetSalesforceAccessToken()

	res, err := conn.CreateMetadata(ctx, input, accessToken)
	if err != nil {
		slog.Error("err", "err", err)
	}

	fmt.Println("Field Operation Result", res)
}
