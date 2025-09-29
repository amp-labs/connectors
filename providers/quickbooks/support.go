package quickbooks

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

//nolint:funlen
func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"account",
		"attachable",
		"bill",
		"billPayment",
		"budget",
		"class",
		"companyCurrency",
		"creditMemo",
		"creditCardPayment",
		"customer",
		"customerType",
		"department",
		"deposit",
		"employee",
		"estimate",
		"exchangeRate",
		"invoice",
		"item",
		"journalCode",
		"journalEntry",
		"payment",
		"paymentMethod",
		"purchase",
		"purchaseOrder",
		"recurringTransaction",
		"refundReceipt",
		"reimburseCharge",
		"salesReceipt",
		"taxCode",
		"taxPayment",
		"taxRate",
		"taxAgency",
		"term",
		"timeActivity",
		"transfer",
		"vendor",
		"vendorCredit",
	}

	writeSupport := []string{
		"account",
		"attachable",
		"bill",
		"billPayment",
		"budget",
		"class",
		"companyCurrency",
		"creditMemo",
		"creditCardPayment",
		"customer",
		"department",
		"deposit",
		"employee",
		"estimate",
		"inventoryadjustment",
		"invoice",
		"item",
		"journalEntry",
		"payment",
		"journalCode",
		"journalEntry",
		"payment",
		"paymentMethod",
		"purchase",
		"purchaseOrder",
		"recurringTransaction",
		"refundReceipt",
		"reimburseCharge",
		"salesReceipt",
		"taxservice/taxcode",
		"taxAgency",
		"term",
		"timeActivity",
		"transfer",
		"vendor",
		"vendorCredit",
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
