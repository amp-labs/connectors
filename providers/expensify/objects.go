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

// Supported object names can be found under schemas.json.
var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNamePolicy,
	objectNameReport,
	objectNameExpenses,
	objectNameExpenseRules,
)
