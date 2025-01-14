package asana

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameAllocation  = "allocations"
	objectNameGoals       = "goals"
	objectNameMemberships = "memberships"
	objectNameProjects    = "projects"
)

var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNameAllocation,
	objectNameGoals,
	objectNameMemberships,
	objectNameProjects,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameAllocation,
	objectNameGoals,
	objectNameMemberships,
	objectNameProjects,
)
