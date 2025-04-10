package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	res, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "ApexClass",
		RecordData: map[string]any{
			"ApiVersion": 57.0,
			"Status":     "Active",
			"Body":       "public class GreetingClassX {public String userName { get; set; }\n\n    public GreetingClassX() {\n\n        userName = 'Guest'; // Default value\n    }\n\n    public String sayHello() {\n        return 'Hello, ' + userName + '!';\n    }\n\n    public void submitName() {\n        // Custom logic to handle the submitted name\n        System.debug('Submitted name: ' + userName);\n    }\n}",
		},
	})
	if err != nil {
		utils.Fail("error reading from Salesforce: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
