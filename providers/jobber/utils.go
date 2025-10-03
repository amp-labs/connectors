package jobber

import "github.com/amp-labs/connectors/internal/datautils"

const apiVersion = "2025-01-20"

// Jobber API Documentation: https://developer.getjobber.com/docs
// This link provides an overview of Jobber API objects.
// The full list of queries and mutations can be retrieved after logging in.
// To explore them, go to "Manage Apps", click "Actions" for the respective app,
// and then click "Test in GraphiQL".
var objectNameMapping = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"appAlerts":                 "AppAlert",
	"apps":                      "Application",
	"capitalLoans":              "JobberPaymentsCpitalLoan",
	"clientEmails":              "Email",
	"clientPhones":              "ClientPhoneNumber",
	"clients":                   "Client",
	"Expenses":                  "Expense",
	"invoices":                  "Invoice",
	"jobs":                      "Job",
	"paymentsRecords":           "PaymentRecordInterface",
	"payoutRecords":             "PayoutRecord",
	"products":                  "ProductOrService",
	"properties":                "Property",
	"quotes":                    "Quote",
	"requestSettingsCollection": "RequestSettings",
	"requests":                  "Request",
	"scheduledItems":            "ScheduledItemInterface",
	"similarClients":            "Client",
	"tasks":                     "Task",
	"taxRates":                  "TaxRate",
	"timeSheetEntries":          "TimeSheetEntry",
	"users":                     "User",
	"vehicles":                  "Vehicle",
	"vists":                     "Visit",
}, func(objectName string) string {
	return objectName
})
