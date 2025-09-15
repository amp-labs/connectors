package jobber

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"appAlerts",
		"apps",
		"capitalLoans",
		"clientEmails",
		"clientPhones",
		"clients",
		"Expenses",
		"invoices",
		"jobs",
		"payoutRecords",
		"products",
		"properties",
		"quotes",
		"requestSettingsCollection",
		"requests",
		"similarClients",
		"tasks",
		"taxRates",
		"timeSheetEntries",
		"users",
		"vehicles",
		"vists",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
