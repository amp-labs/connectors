package support

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
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
		// Alternative endpoints are used instead.
		// Each endpoint has a corresponding endpoint supporting incremental read.
		"/api/v2/organizations",       // => /api/v2/incremental/organizations
		"/api/v2/routing/attributes",  // => /api/v2/incremental/routing/attributes
		"/api/v2/tickets",             // => /api/v2/incremental/tickets/cursor
		"/api/v2/users",               // => /api/v2/incremental/users/cursor
		"/api/v2/incremental/tickets", // cursor is preferred over raw incremental
		"/api/v2/incremental/users",   // cursor is preferred over raw incremental
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/api/v2/incremental/routing/attributes": "attributes",
		"/api/v2/incremental/tickets/cursor":     "tickets",
		"/api/v2/incremental/users/cursor":       "users",
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"search":               "Search Results",
		"trigger_categories":   "Trigger Categories",
		"satisfaction_reasons": "Satisfaction Rating Reasons",
		"ticket_audits":        "Ticket Audits",
	}
	objectNameToResponseField = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
		"ticket_audits":        "audits",
		"search":               "results", // This is "/api/v2/search"
		"satisfaction_reasons": "reasons",
	}, func(objectName string) (fieldName string) {
		return objectName
	})
	objectNameToPagination = map[string]string{ // nolint:gochecknoglobals
		"activities":                 "cursor",
		"attribute_values":           "time",
		"attributes":                 "time",
		"audit_logs":                 "cursor",
		"automations":                "cursor",
		"bookmarks":                  "offset",
		"brands":                     "cursor",
		"custom_objects":             "offset",
		"custom_roles":               "offset",
		"custom_statuses":            "",
		"deleted_tickets":            "cursor",
		"deleted_users":              "cursor",
		"deletion_schedules":         "",
		"email_notifications":        "cursor",
		"group_memberships":          "cursor",
		"groups":                     "cursor",
		"instance_values":            "time",
		"job_statuses":               "cursor",
		"locales":                    "offset",
		"macros":                     "cursor",
		"monitored_twitter_handles":  "offset",
		"organization_fields":        "cursor",
		"organization_memberships":   "cursor",
		"organization_subscriptions": "cursor",
		"organizations":              "time",
		"queues":                     "offset",
		"recipient_addresses":        "cursor",
		"requests":                   "cursor",
		"resource_collections":       "offset",
		"satisfaction_ratings":       "cursor",
		"satisfaction_reasons":       "offset",
		"search":                     "offset",
		"session":                    "",
		"sessions":                   "",
		"sharing_agreements":         "offset",
		"suspended_tickets":          "cursor",
		"tags":                       "cursor",
		"target_failures":            "offset",
		"targets":                    "offset",
		"ticket_audits":              "cursor",
		"ticket_events":              "time",
		"ticket_fields":              "cursor",
		"ticket_forms":               "offset",
		"ticket_metric_events":       "time",
		"ticket_metrics":             "cursor",
		"tickets":                    "cursor",
		"trigger_categories":         "cursor",
		"triggers":                   "cursor",
		"user_fields":                "cursor",
		"users":                      "cursor",
		"views":                      "cursor",
		"workspaces":                 "offset",
	}
	objectNameIncremental = datautils.NewSet( // nolint:gochecknoglobals
		"attribute_values",
		"attributes",
		"instance_values",
		"organizations",
		"ticket_events",
		"ticket_metric_events",
		"tickets",
		"users",
	)
)

func Objects() []metadatadef.ExtendedSchema[metadata.CustomProperties] {
	explorer, err := openapi.SupportFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverride,
		api3.CustomMappingObjectCheck(objectNameToResponseField),
	)
	goutils.MustBeNil(err)

	for index, object := range objects {
		object.Custom.Pagination = objectNameToPagination[object.ObjectName]
		object.Custom.Incremental = objectNameIncremental.Has(object.ObjectName)

		objects[index] = object
	}

	return objects
}
