package aha

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/aha/metadata"
)

const (
	objectNameAudits             = "audits"
	objectNameHistoricalAudits   = "historical_audits"
	objectNameIdeasEndorsements  = "ideas/endorsements"
	objectNameIdeas              = "ideas"
	objectNameIdeaOrganization   = "idea_organizations"
	objectNameIdeaUser           = "idea_users"
	objectNameIntegrations       = "integrations"
	objectNameProducts           = "products"
	objectNameReleasePhases      = "release_phases"
	objectNmaeSchedulableChanges = "schedulable_changes"
	objectNameTeamMembers        = "team_members"
	objectNameTeams              = "teams"
	objectNameTasks              = "tasks"
)

var supportSince = datautils.NewSet( //nolint:gochecknoglobals
	objectNameAudits,
	objectNameHistoricalAudits,
	objectNameIdeasEndorsements,
	objectNameIdeas,
)

var supportWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameHistoricalAudits,
	objectNameIdeaOrganization,
	objectNameIdeaUser,
	objectNameIntegrations,
	objectNameProducts,
	objectNameReleasePhases,
	objectNameTeamMembers,
	objectNameTasks,
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(staticschema.RootModuleID)

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportWrite.List(), ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
