package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/amp-labs/connectors/zendesksupport/metadata"
	"github.com/amp-labs/connectors/zendesksupport/openapi"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		// Wild rules.
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
		// Complex Path.
		"/api/v2/channels/twitter/tickets",
		"/api/v2/suspended_tickets/attachments",
		"/api/v2/group_memberships/assignable",
		"/api/v2/dynamic_content/items",
		"/api/v2/slas/policies",
		"/api/v2/macros/*",
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
	}
	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		// Path: /api/v2/problems -> additionalProperties (this is a dictionary/map/free-form data structure)
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"search":               "Search Results",
		"trigger_categories":   "Trigger Categories",
		"satisfaction_reasons": "Satisfaction Rating Reasons",
		"ticket_audits":        "Ticket Audits",
	}
	objectNameToResponseField = map[string]string{ // nolint:gochecknoglobals
		"ticket_audits":        "audits",
		"search":               "results", // This is "/api/v2/search"
		"satisfaction_reasons": "reasons",
	}
)

func main() {
	explorer, err := openapi.FileManager.GetExplorer()
	must(err)

	objects, err := explorer.GetBasicReadObjects(
		ignoreEndpoints, objectEndpoints, displayNameOverride, IsResponseFieldAppropriate,
	)
	must(err)

	schemas := scrapper.NewObjectMetadataResult()

	for _, object := range objects {
		for _, field := range object.Fields {
			schemas.Add(object.ObjectName, object.DisplayName, field, nil)
		}
	}

	must(metadata.FileManager.SaveSchemas(schemas))

	slog.Info("Completed.")
}

func IsResponseFieldAppropriate(fieldName, objectName string) bool {
	if responseFieldName, ok := objectNameToResponseField[objectName]; ok {
		return fieldName == responseFieldName
	}

	return api3.IdenticalObjectCheck(fieldName, objectName)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
