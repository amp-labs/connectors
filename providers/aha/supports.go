package aha

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
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
	objectNameSchedulableChanges = "schedulable_changes"
	objectNameTeamMembers        = "team_members"
	objectNameTeams              = "teams"
	objectNameTasks              = "tasks"
)

var supportSince = datautils.NewSet( //nolint:gochecknoglobals
	// https://www.aha.io/api/resources/audits/retrieve_record_history
	objectNameAudits,
	// https://www.aha.io/api/resources/historical_audits/create_an_audit_search
	objectNameHistoricalAudits,
	// https://www.aha.io/api/resources/idea_votes/list_votes_for_an_account
	objectNameIdeasEndorsements,
	// https://www.aha.io/api/resources/ideas/list_ideas
	objectNameIdeas,
)

var supportUntil = datautils.NewSet( //nolint:gochecknoglobals
	// https://www.aha.io/api/resources/audits/retrieve_record_history
	objectNameAudits,
	// https://www.aha.io/api/resources/historical_audits/create_an_audit_search
	objectNameHistoricalAudits,
	// https://www.aha.io/api/resources/idea_votes/list_votes_for_an_account
	objectNameIdeasEndorsements,
	// https://www.aha.io/api/resources/ideas/list_ideas
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
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
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
