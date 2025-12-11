package jobber

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "2025-01-20"
	defaultPageSize = 50
)

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

func makeNextRecordsURL(objName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node, "data", objName).ObjectOptional("pageInfo")
		if err != nil {
			return "", err
		}

		if pagination != nil {
			hasNextPage, err := jsonquery.New(pagination).BoolOptional("hasNextPage")
			if err != nil {
				return "", err
			}

			if !(*hasNextPage) {
				return "", nil
			}

			endCursorToken, err := jsonquery.New(pagination).StringOptional("endCursor")
			if err != nil {
				return "", err
			}

			return *endCursorToken, nil
		}

		return "", nil
	}
}

// Singularize all objectname expect productsAndServices object.
func getObjectName(objName string) string {
	if objName == "productsAndServices" {
		return objName
	}

	return naming.NewSingularString(objName).String()
}

// All write objects use the singular form of the node path, except for the productsAndServices object.
var writeObjectNodePathMapping = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"productsAndServices": "productOrService",
}, func(objectName string) string {
	return naming.NewSingularString(objectName).String()
})
