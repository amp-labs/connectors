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
		objectCapitalLoans,
		"clientEmails",
		"clientPhones",
		objectClients,
		objectExpenses,
		objectInvoices,
		objectJobs,
		objectPayoutRecords,
		objectProducts,
		objectProperties,
		objectQuotes,
		"requestSettingsCollection",
		objectRequests,
		"similarClients",
		objectTasks,
		"taxRates",
		objectTimeSheetEntries,
		objectUsers,
		objectVehicles,
		objectVisits,
	}

	writeSupport := []string{
		objectClients,
		"events",
		objectExpenses,
		objectJobs,
		objectProductsAndServices,
		objectQuotes,
		objectRequests,
		"taxes",
		"taxGroups",
		objectVehicles,
	}

	deleteSupport := []string{
		objectExpenses,
		objectVehicles,
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
