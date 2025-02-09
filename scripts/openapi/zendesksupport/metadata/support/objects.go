package support

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/providers/zendesksupport/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		// Wild rules.
		"/api/lotus/*",
		"*/create_many",
		"*/update_many",
		"*/destroy_many",
		"*/reorder",
		"*/count",
		"*/create_or_update",
		"*/show_many",
		"/api/v2/incremental/*",
		"/api/v2/autocomplete/*",
		"*/autocomplete",
		"*/active",
		"*/export",
		"*/definitions",
		"*/assignable",
		// Complex Path with multiple slashes.
		"/api/v2/channels/twitter/tickets",
		"/api/v2/suspended_tickets/attachments",
		"/api/v2/dynamic_content/items",
		"/api/v2/slas/policies",
		"/api/v2/macros/*",
		"/api/v2/object_layouts/essentials_cards",
		"/api/v2/locales/public",
		"/api/v2/views/compact",
		"/api/v2/locales/agent",
		"/api/v2/group_slas/policies",
		"/api/v2/slas/policies",
		"/api/v2/routing/requirements/fulfilled",
		// Resources with search.
		"/api/v2/users/search",
		"/api/v2/requests/search",
		"/api/v2/organizations/search",
		"/api/v2/automations/search",
		"/api/v2/views/search",
		"/api/v2/triggers/search",
		// Not applicable.
		"/api/v2/channels/voice/tickets", // only POST method for create.
		"/api/v2/imports/tickets",        // only POST method for create.
		"/api/v2/custom_objects/limits/object_limit",
		"/api/v2/users/me/session/renew",
		"/api/v2/locales/current",
		"/api/v2/locales/detect_best_locale",
		"/api/v2/brands/check_host_mapping",
		"/api/v2/views/count_many",
		"/api/v2/accounts/available",
		"/api/v2/users/me",
		"/api/v2/custom_objects/limits/record_limit",
		"/api/v2/account/settings",
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"search":               "Search Results",
		"trigger_categories":   "Trigger Categories",
		"satisfaction_reasons": "Satisfaction Rating Reasons",
		"ticket_audits":        "Ticket Audits",
	}
)

func Objects() []api3.Schema {
	explorer, err := openapi.SupportFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, displayNameOverride,
		api3.CustomMappingObjectCheck(zendesksupport.ObjectNameToResponseField[zendesksupport.ModuleTicketing]),
	)
	goutils.MustBeNil(err)

	return objects
}
