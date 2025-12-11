package quickbooks

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

var objectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"creditCardPayment": "CreditCardPaymentTxn",
},
	func(objectName string) string {
		return objectName
	},
)

//nolint:funlen
func supportedOperations() components.EndpointRegistryInput {
	//nolint:lll,gofumpt
	readSupport := []string{
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/account#query-an-account
		"account",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/attachable#query-an-attachable
		"attachable",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/bill#query-a-bill
		"bill",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/billpayment#query-a-billpayment
		"billPayment",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/budget#query-a-budget
		"budget",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/class#query-a-class
		"class",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/companycurrency#query-a-companycurrency
		"companyCurrency",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/creditmemo#query-a-credit-memo
		"creditMemo",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/creditcardpayment#query-a-creditcardpayment
		"creditCardPayment",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/customer#query-a-customer
		"customer",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/customertype#query-a-customertype
		"customerType",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/department#query-a-department
		"department",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/deposit#query-a-deposit
		"deposit",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/employee#query-an-employee
		"employee",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/estimate#query-an-estimate
		"estimate",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/exchangerate#query-exchangerate-objects.
		"exchangeRate",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/invoice#query-an-invoice
		"invoice",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/item#query-an-item
		"item",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/journalcode#query-a-journalcode
		"journalCode",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/journalentry#query-a-journalentry
		"journalEntry",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/payment#query-a-payment
		"payment",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/paymentmethod#query-a-paymentmethod
		"paymentMethod",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/purchase#query-a-purchase
		"purchase",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/purchaseorder#query-a-purchaseorder
		"purchaseOrder",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/recurringtransaction#query-a-recurring-transaction
		"recurringTransaction",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/refundreceipt#query-a-refundreceipt
		"refundReceipt",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/reimbursecharge#query-a-reimbursecharge
		"reimburseCharge",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/salesreceipt#query-a-salesreceipt
		"salesReceipt",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/taxcode#query-a-taxcode
		"taxCode",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/taxpayment#query-taxpayment
		"taxPayment",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/taxrate#query-a-taxrate
		"taxRate",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/taxagency#query-a-taxagency
		"taxAgency",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/term#query-a-term
		"term",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/timeactivity#query-a-timeactivity-object
		"timeActivity",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/transfer#query-a-transfer
		"transfer",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/vendor#query-a-vendor
		"vendor",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/vendorcredit#query-a-vendorcredit
		"vendorCredit",
	}

	//nolint:lll,gofumpt
	writeSupport := []string{
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/account#create-an-account
		"account",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/attachable#create-a-note-attachment
		"attachable",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/bill#create-a-bill
		"bill",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/billpayment#create-a-billpayment
		"billPayment",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/budget#create-a-budget
		"budget",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/class#create-a-class
		"class",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/companycurrency#create-a-companycurrency
		"companyCurrency",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/creditmemo#create-a-credit-memo
		"creditMemo",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/creditcardpayment#create-a-creditcardpayment
		"creditCardPayment",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/customer#create-a-customer
		"customer",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/department#create-a-department
		"department",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/deposit#create-a-deposit
		"deposit",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/employee#create-an-employee
		"employee",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/estimate#create-an-Estimate
		"estimate",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/inventoryadjustment#create-an-inventory-adjustment
		"inventoryadjustment",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/invoicing#create-an-invoice
		"invoice",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/item#create-an-item
		"item",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/journalcode#create-a-journalcode
		"journalCode",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/journalentry#create-a-journalentry
		"journalEntry",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/payment#create-a-payment
		"payment",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/paymentmethod#create-a-paymentmethod
		"paymentMethod",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/purchase
		"purchase",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/purchaseorder#create-a-purchaseorder
		"purchaseOrder",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/recurringtransaction#create-a-recurring-transaction
		"recurringTransaction",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/refundreceipt#create-a-refund-receipt
		"refundReceipt",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/salesreceipt#create-a-salesreceipt
		"salesReceipt",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/taxservice#create-a-taxservice
		"taxservice/taxcode",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/taxagency#create-a-taxagency
		"taxAgency",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/term#create-a-term
		"term",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/timeactivity#create-a-timeactivity-object
		"timeActivity",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/transfer#create-an-transfer
		"transfer",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/vendor#create-a-vendor
		"vendor",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/vendorcredit#create-a-vendorcredit
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
