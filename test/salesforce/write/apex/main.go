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
		RecordData: `<?xml version="1.0" encoding="UTF-8"?>
<ApexClass xmlns="http://soap.sforce.com/2006/04/metadata">
    <apiVersion>57.0</apiVersion>
    <status>Active</status>
    <body>
        public class GreetingClass {
        public String userName { get; set; }

        public GreetingClass() {
        userName = 'Guest'; // Default value
        }

        public String sayHello() {
        return 'Hello, ' + userName + '!';
        }

        public void submitName() {
        // Custom logic to handle the submitted name
        System.debug('Submitted name: ' + userName);
        }
        }
    </body>
</ApexClass>`,
	})
	if err != nil {
		utils.Fail("error reading from Salesforce: %w", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
