package asana

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/asana/metadata"
)

// Some of the objects (allocations, goals, memberships, portfolios, tasks)
// require us to pass either the team ID or the workspace.
// although the API documentation doesn’t explicitly state that these fields are mandatory for fetching data, they are.

const (
	objectNameAccessRequests      = "access_requests"
	objectNameAllocations         = "allocations"
	objectNameCustomFields        = "custom_fields"
	objectNameGoals               = "goals"
	objectNameMemberships         = "memberships"
	objectNameOrganizationExports = "organization_exports"
	objectNamePortfolios          = "portfolios"
	objectNameProjects            = "projects"
	objectNameStatusUpdates       = "status_updates"
	objectNameTags                = "tags"
	objectNameTasks               = "tasks"
	objectNameTeams               = "teams"
	objectNameUsers               = "users"
	objectNameWebhooks            = "webhooks"
	objectNameWorkspaces          = "workspaces"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameAccessRequests,
	objectNameAllocations,
	objectNameCustomFields,
	objectNameGoals,
	objectNameMemberships,
	objectNameOrganizationExports,
	objectNamePortfolios,
	objectNameProjects,
	objectNameStatusUpdates,
	objectNameTags,
	objectNameTasks,
	objectNameTeams,
	objectNameWebhooks,
)

var supportLimitAndOffset = datautils.NewSet( //nolint:gochecknoglobals
	objectNameWorkspaces,
	objectNameUsers,
	objectNameProjects,
	objectNameTags,
	objectNameWorkspaces,
)
