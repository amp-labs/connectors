package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
	"github.com/amp-labs/connectors/scripts/openapi/zendesksupport/metadata/helpcenter"
	"github.com/amp-labs/connectors/scripts/openapi/zendesksupport/metadata/support"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func main() {
	schemas := staticschema.NewExtendedMetadata[staticschema.FieldMetadataMapV1, metadata.CustomProperties]()
	registry := datautils.NamedLists[string]{}
	lists := datautils.IndexedLists[common.ModuleID, metadatadef.Schema]{}

	lists.Add(zendesksupport.ModuleTicketing, support.Objects()...)
	lists.Add(zendesksupport.ModuleHelpCenter, helpcenter.Objects()...)

	for module, objects := range lists {
		for _, object := range objects {
			if object.Problem != nil {
				slog.Error("schema not extracted",
					"objectName", object.ObjectName,
					"error", object.Problem,
				)
			}

			for _, field := range object.Fields {
				schemas.Add(module, object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
					staticschema.FieldMetadataMapV1{
						field.Name: field.Name,
					}, nil, metadata.CustomProperties{
						Pagination: findPagination(module, object.ObjectName),
					})
			}

			for _, queryParam := range object.QueryParams {
				registry.Add(queryParam, object.ObjectName)
			}
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.")
}

func findPagination(module common.ModuleID, name string) string { // nolint:funlen
	switch module {
	case zendesksupport.ModuleTicketing:
		return map[string]string{
			"attributes":                 "offset",
			"audit_logs":                 "cursor",
			"organizations":              "cursor",
			"activities":                 "cursor",
			"automations":                "cursor",
			"deleted_users":              "cursor",
			"deletion_schedules":         "",
			"monitored_twitter_handles":  "offset",
			"target_failures":            "offset",
			"bookmarks":                  "offset",
			"group_memberships":          "cursor",
			"locales":                    "offset",
			"ticket_fields":              "cursor",
			"workspaces":                 "offset",
			"deleted_tickets":            "cursor",
			"queues":                     "offset",
			"ticket_forms":               "offset",
			"triggers":                   "cursor",
			"user_fields":                "cursor",
			"views":                      "cursor",
			"organization_memberships":   "cursor",
			"requests":                   "cursor",
			"suspended_tickets":          "cursor",
			"ticket_audits":              "cursor",
			"trigger_categories":         "cursor",
			"session":                    "",
			"targets":                    "offset",
			"email_notifications":        "cursor",
			"groups":                     "cursor",
			"job_statuses":               "cursor",
			"recipient_addresses":        "cursor",
			"satisfaction_reasons":       "offset",
			"search":                     "offset",
			"ticket_metrics":             "cursor",
			"tickets":                    "cursor",
			"custom_objects":             "offset",
			"organization_fields":        "cursor",
			"organization_subscriptions": "cursor",
			"resource_collections":       "offset",
			"satisfaction_ratings":       "cursor",
			"sessions":                   "",
			"users":                      "cursor",
			"brands":                     "cursor",
			"custom_roles":               "offset",
			"custom_statuses":            "",
			"macros":                     "cursor",
			"sharing_agreements":         "offset",
			"tags":                       "cursor",
		}[name]
	case zendesksupport.ModuleHelpCenter:
		return map[string]string{
			"posts":           "cursor",
			"topics":          "cursor",
			"user_segments":   "cursor",
			"article_labels":  "cursor",
			"articles":        "offset",
			"community_posts": "offset",
		}[name]
	}

	return ""
}
