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
		"expenses",
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
		"visits",
	}

	writeSupport := []string{
		"clients",
		"events",
		"expenses",
		"jobs",
		"productsAndServices",
		"quotes",
		"requests",
		"taxes",
		"taxGroups",
		"vehicles",
	}

	deleteSupport := []string{
		"expenses",
		"vehicles",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(deleteSupport, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
