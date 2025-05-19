package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/aws"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAWSConnector(ctx, providers.ModuleAWSIdentityCenter)

	objects := []string{
		"Instances",
		"Applications",
		"ApplicationProviders",
		"AccountAssignmentCreationStatus",
		"AccountAssignmentDeletionStatus",
		"PermissionSetProvisioningStatus",
		"TrustedTokenIssuers",
		"Groups",
		"Users",
	}

	for _, objectName := range objects {
		res, err := conn.Read(ctx, common.ReadParams{
			ObjectName: objectName,
			Fields:     connectors.Fields("id"),
		})
		if err != nil {
			fmt.Println("error", err.Error(), "object", objectName)

			continue
		}

		if len(res.Data) == 0 {
			fmt.Println("Empty output", objectName)
		} else {
			fmt.Println("✅", objectName)
		}
	}
}
