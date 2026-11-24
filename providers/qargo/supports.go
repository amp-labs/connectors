package qargo

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

const (
	apiVersion = "v1"

	// Companies (GET /v1/companies/company). A company can be a customer, a subcontractor and a
	// supplier at once; customers are the companies with is_customer set.
	// https://api-docs.qargo.com/docs/section/concepts/company
	objectNameCompany = "company"

	// Approved sales invoices and credit notes, read from the accounting sync-task queue (see queue.go).
	objectNameSalesInvoice    = "sales-invoice"
	objectNameSalesCreditNote = "sales-credit-note"

	// Payment status of a sales invoice, keyed by the invoice id:
	// POST /v1/accounting/sales-invoice/{id}/payment/update-status. Update only; there is no create.
	// https://api-docs.qargo.com/docs/api-accounting ("Report Payment Status For Sales Invoice")
	objectNameSalesInvoicePayment = "sales-invoice/payment"
)

var readObjects = []string{ // nolint:gochecknoglobals
	objectNameCompany,
	objectNameSalesInvoice,
	objectNameSalesCreditNote,
}

// Sync task type that carries each queue object, POST_INVOICE or POST_CREDIT_NOTE (TaskType in the spec).
var queueTaskTypes = map[string]string{ // nolint:gochecknoglobals
	objectNameSalesInvoice:    "POST_INVOICE",
	objectNameSalesCreditNote: "POST_CREDIT_NOTE",
}

// SupportedReadObjects lists the objects this connector reads.
func SupportedReadObjects() []string {
	return slices.Clone(readObjects)
}

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readObjects, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: objectNameSalesInvoicePayment,
				Support:  components.WriteSupport,
			},
		},
	}
}
