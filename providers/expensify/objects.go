package expensify

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	objectNamePolicy       = "policy"
	objectNameReport       = "report"
	objectNameExpenses     = "expenses"
	objectNameExpenseRules = "expenseRules"
)

// ref: https://integrations.expensify.com/Integration-Server/doc/#read-get
var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNamePolicy,
)

var readObjectResponseIdentifier = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNamePolicy: "policyList",
},
	func(objectName string) string {
		return objectName
	},
)

// ref; https://integrations.expensify.com/Integration-Server/doc/#create
var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNamePolicy,
	objectNameReport,
	objectNameExpenses,
	objectNameExpenseRules,
)
