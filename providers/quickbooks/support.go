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

	//nolint:lll,gofumpt
	writeSupport := []string{
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/account#create-an-account
		"account",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/attachable#create-a-note-attachment
		"attachable",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/bill#create-a-bill
		"bill",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/billpayment#create-a-billpayment
		"billPayment",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/budget#create-a-budget
		"budget",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/class#create-a-class
		"class",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/companycurrency#create-a-companycurrency
		"companyCurrency",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/creditmemo#create-a-credit-memo
		"creditMemo",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/creditcardpayment#create-a-creditcardpayment
		"creditCardPayment",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/customer#create-a-customer
		"customer",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/department#create-a-department
		"department",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/deposit#create-a-deposit
		"deposit",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/employee#create-an-employee
		"employee",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/estimate#create-an-Estimate
		"estimate",
		// https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/inventoryadjustment#create-an-inventory-adjustment
		"inventoryadjustment",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/invoicing#create-an-invoice
		"invoice",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/item#create-an-item
		"item",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/journalcode#create-a-journalcode
		"journalCode",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/journalentry#create-a-journalentry
		"journalEntry",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/payment#create-a-payment
		"payment",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/paymentmethod#create-a-paymentmethod
		"paymentMethod",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/purchase
		"purchase",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/purchaseorder#create-a-purchaseorder
		"purchaseOrder",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/recurringtransaction#create-a-recurring-transaction
		"recurringTransaction",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/refundreceipt#create-a-refund-receipt
		"refundReceipt",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/salesreceipt#create-a-salesreceipt
		"salesReceipt",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/taxservice#create-a-taxservice
		"taxservice/taxcode",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/taxagency#create-a-taxagency
		"taxAgency",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/term#create-a-term
		"term",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/timeactivity#create-a-timeactivity-object
		"timeActivity",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/transfer#create-an-transfer
		"transfer",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/vendor#create-a-vendor
		"vendor",
		//https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/vendorcredit#create-a-vendorcredit
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
