package xero

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"accounts",
		"contacts",
		"bankTransactions",
		"bankTransfers",
		"batchPayments",
		"brandingThemes",
		"budgets",
		"contactGroups",
		"creditNotes",
		"currencies",
		"invoices",
		"items",
		"journals",
		"linkedTransactions",
		"manualJournals",
		"organisation",
		"overpayments",
		"paymentServices",
		"payments",
		"prepayments",
		"purchaseOrders",
		"quotes",
		"repeatingInvoices",
		"reports",
		"taxRates",
		"trackingCategories",
		"users",
	}

	writeSupport := []string{
		"bankTransactions",
		"bankTransfers",
		"contactGroups",
		"trackingCategories",
		"taxRates",
		"quotes",
		"purchaseOrders",
		"paymentServices",
		"manualJournals",
		"linkedTransactions",
		"items",
		"invoices",
		"currencies",
		"creditNotes",
		"contacts",
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
		},
	}
}
