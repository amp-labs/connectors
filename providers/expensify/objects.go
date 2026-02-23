package expensify

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// Some of the objects (allocations, goals, memberships, portfolios, tasks)
// require us to pass either the team ID or the workspace.
// although the API documentation doesn’t explicitly state that these fields are mandatory for fetching data, they are.

const (
	objectNamePolicy       = "policy"
	objectNameReport       = "report"
	objectNameExpenses     = "expenses"
	objectNameExpenseRules = "expenseRules"
)

// Supported object names can be found under schemas.json.
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
