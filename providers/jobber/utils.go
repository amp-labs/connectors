package jobber

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// apiVersion pins the Jobber GraphQL schema version.
	// Before bumping it, verify the embedded queries and metadata still work:
	// Jobber can remove fields even from already-served versions,
	// and one unknown field fails an entire query.
	apiVersion      = "2026-05-12"
	defaultPageSize = 50

	objectCapitalLoans        = "capitalLoans"
	objectClients             = "clients"
	objectExpenses            = "expenses"
	objectInvoices            = "invoices"
	objectJobs                = "jobs"
	objectPayoutRecords       = "payoutRecords"
	objectProducts            = "products"
	objectProperties          = "properties"
	objectQuotes              = "quotes"
	objectRequests            = "requests"
	objectTasks               = "tasks"
	objectTimeSheetEntries    = "timeSheetEntries"
	objectUsers               = "users"
	objectVehicles            = "vehicles"
	objectVisits              = "visits"
	objectProductsAndServices = "productsAndServices"

	// Timestamp fields used for incremental read.
	fieldUpdatedAt = "updatedAt"
	fieldCreatedAt = "createdAt"

	// GraphQL request body keys.
	gqlQueryKey     = "query"
	gqlVariablesKey = "variables"
)

// Jobber API Documentation: https://developer.getjobber.com/docs
// This link provides an overview of Jobber API objects.
// The full list of queries and mutations can be retrieved after logging in.
// To explore them, go to "Manage Apps", click "Actions" for the respective app,
// and then click "Test in GraphiQL".
var objectNameMapping = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"appAlerts":                 "AppAlert",
	"apps":                      "Application",
	objectCapitalLoans:          "JobberPaymentsCapitalLoan",
	"clientEmails":              "Email",
	"clientPhones":              "ClientPhoneNumber",
	objectClients:               "Client",
	objectExpenses:              "Expense",
	objectInvoices:              "Invoice",
	objectJobs:                  "Job",
	"paymentsRecords":           "PaymentRecordInterface",
	objectPayoutRecords:         "PayoutRecord",
	objectProducts:              "ProductOrService",
	objectProperties:            "Property",
	objectQuotes:                "Quote",
	"requestSettingsCollection": "RequestSettings",
	objectRequests:              "Request",
	"scheduledItems":            "ScheduledItemInterface",
	"similarClients":            "Client",
	objectTasks:                 "Task",
	"taxRates":                  "TaxRate",
	objectTimeSheetEntries:      "TimeSheetEntry",
	objectUsers:                 "User",
	objectVehicles:              "Vehicle",
	objectVisits:                "Visit",
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
	if objName == objectProductsAndServices {
		return objName
	}

	return naming.NewSingularString(objName).String()
}

// All write objects use the singular form of the node path, except for the productsAndServices object.
var writeObjectNodePathMapping = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectProductsAndServices: "productOrService",
}, func(objectName string) string {
	return naming.NewSingularString(objectName).String()
})
