package exensify

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// Some of the objects (allocations, goals, memberships, portfolios, tasks)
// require us to pass either the team ID or the workspace.
// although the API documentation doesn’t explicitly state that these fields are mandatory for fetching data, they are.

const (
	objectNamePolicy = "policy"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNamePolicy,
)
