package loxo

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput { // nolint:funlen
	readSupport := []string{
		"activity_types",
		"address_types",
		"bonus_payment_types",
		"bonus_types",
		"companies",
		"company_global_statuses",
		"company_types",
		"compensation_types",
		"countries",
		"currencies",
		"deal_workflows",
		"deals",
		"disability_statuses",
		"diversity_types",
		"dynamic_fields",
		"education_types",
		"email_tracking",
		"email_types",
		"equity_types",
		"ethnicities",
		"fee_types",
		"form_templates",
		"forms",
		"genders",
		"job_categories",
		"job_contact_types",
		"job_owner_types",
		"job_statuses",
		"job_types",
		"jobs",
		"people",
		"people/emails",
		"person_phones",
		"people/update_by_email",
		"person_events",
		"person_global_statuses",
		"person_lists",
		"person_share_field_types",
		"person_types",
		"phone_types",
		"placements",
		"pronouns",
		"question_types",
		"schedule_items",
		"scorecards",
		"scorecard_recommendation_types",
		"scorecard_types",
		"scorecard_visibility_types",
		"seniority_levels",
		"sms",
		"social_profile_types",
		"source_types",
		"users",
		"veteran_statuses",
		"workflow_stages",
		"workflows",
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
