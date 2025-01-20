package asana

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameAllocations         = "allocations"
	objectNameGoals               = "goals"
	objectNameMemberships         = "memberships"
	objectNameProjects            = "projects"
	objectNameOrganizationExports = "organization_exports"
	objectNamePortfolios          = "portfolios"
	objectNameStatusUpdates       = "status_updates"
	objectNameTags                = "tags"
	objectNameTasks               = "tasks"
	objectNameTeams               = "teams"
	objectNameUsers               = "users"
	objectNameWorkspaces          = "workspaces"
)

var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	objectNameAllocations,
	objectNameGoals,
	objectNameMemberships,
	objectNameProjects,
	objectNameUsers,
	objectNameWorkspaces,
)

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameAllocations,
	objectNameGoals,
	objectNameMemberships,
	objectNameOrganizationExports,
	objectNamePortfolios,
	objectNameProjects,
	objectNameStatusUpdates,
	objectNameTasks,
	objectNameTeams,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameAllocations,
	objectNameGoals,
	objectNameMemberships,
	objectNameProjects,
	objectNameUsers,
	objectNameWorkspaces,
)
