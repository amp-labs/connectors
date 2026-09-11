// Generates providers/mailgun/metadata/schemas.json from the Mailgun OpenAPI spec.
//
// Run from repo root:
//
//	go run ./scripts/openapi/mailgun/metadata
//
// The Mailgun spec is a merged, multi-version (v1..v5) bundle of several internal
// services, so objects are curated explicitly (allow-list) rather than auto-discovered:
//   - Only genuine list resources scoped by at most the domain are included.
//   - Get-by-id singletons, aggregated analytics (legacy /v3/stats/* and modern
//     /v1/analytics/metrics), action endpoints, reference lookups, pagination
//     variants and sandbox-only endpoints are excluded.
//   - The modern Logs API (POST /v1/analytics/logs) replaces the deprecated GET
//     events endpoint and is read via POST.
//
// Domain-scoped paths keep their placeholder ({domain_name}/{domain}/{name}/
// {authority_name}); the read layer substitutes it from connector metadata.
package main

import (
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
	"github.com/amp-labs/connectors/providers/mailgun/metadata/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// getEndpoints maps GET list URL paths → ObjectName. Domain-scoped placeholders
// are kept verbatim; the read connector substitutes them at request time.
//
//nolint:gochecknoglobals
var getEndpoints = map[string]string{
	// Domain-scoped (domain supplied via connector metadata).
	"/v3/{domain_name}/bounces":             "bounces",
	"/v3/{domain_name}/complaints":          "complaints",
	"/v3/{domain_name}/unsubscribes":        "unsubscribes",
	"/v3/{domain_name}/whitelists":          "whitelists",
	"/v3/{domain}/tags":                     "tags",
	"/v3/domains/{domain_name}/credentials": "domains/credentials",
	"/v3/{domain_name}/templates":           "templates",
	"/v4/domains/{authority_name}/keys":     "domains/keys",

	// Account-scoped.
	"/v4/domains":                           "domains",
	"/v3/routes":                            "routes",
	"/v3/lists":                             "lists",
	"/v3/lists/{list_address}/members":      "lists/members",
	"/v3/forwards":                          "forwards",
	"/v3/ip_pools":                          "ip_pools",
	"/v3/ips":                               "ips",
	"/v1/keys":                              "keys",
	"/v1/dkim/keys":                         "dkim/keys",
	"/v1/webhooks":                          "webhooks",
	"/v5/accounts/subaccounts":              "accounts/subaccounts",
	"/v5/users":                             "users",
	"/v2/ip_whitelist":                      "ip_whitelist",
	"/v1/dynamic_pools/domains":             "dynamic_pools/domains",
	"/v1/thresholds/limits":                 "thresholds/limits",
	"/v1/thresholds/alerts/send":            "thresholds/alerts/send",
	"/v5/accounts/subaccounts/ip_pools/all": "accounts/subaccounts/ip_pools/all",
	// Account-level templates collide with domain templates after truncation
	// (both → "templates"); no distinguishing real path segment exists, so this
	// is a documented naming exception. See metadata/README.md.
	"/v4/templates": "account/templates",
}

// postEndpoints maps POST list URL paths → ObjectName (read via POST body).
//
//nolint:gochecknoglobals
var postEndpoints = map[string]string{
	"/v1/analytics/logs": "analytics/logs",
}

//nolint:gochecknoglobals
var displayNameOverride = map[string]string{
	"ips":                               "IPs",
	"ip_pools":                          "IP Pools",
	"ip_whitelist":                      "IP Allowlist",
	"dkim/keys":                         "DKIM Keys",
	"domains/keys":                      "Domain Keys",
	"domains/credentials":               "Domain Credentials",
	"lists/members":                     "Mailing List Members",
	"lists":                             "Mailing Lists",
	"analytics/logs":                    "Logs",
	"thresholds/limits":                 "Limits",
	"thresholds/alerts/send":            "Send Alerts",
	"account/templates":                 "Account Templates",
	"accounts/subaccounts":              "Subaccounts",
	"accounts/subaccounts/ip_pools/all": "Delegated IP Pools",
	"dynamic_pools/domains":             "Dynamic IP Pool Domains",
}

func pathsOf(m map[string]string) []string {
	paths := make([]string, 0, len(m))
	for path := range m {
		paths = append(paths, path)
	}

	return paths
}

func main() {
	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	objects := Objects()

	for _, object := range objects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"urlPath", object.URLPath,
				"error", object.Problem,
			)

			continue
		}

		for _, field := range object.Fields {
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName,
				object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	slog.Info("Completed.", "objects", len(objects))
}

// Objects extracts schemas for curated GET list endpoints and the POST Logs
// endpoint. ReadObjects is used (not ReadObjectsGet) so single-level {id} paths
// — the {domain} scope — are kept; NestedIDPathIgnorer rejects 2-level nesting.
func Objects() []metadatadef.Schema {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithArrayItemAutoSelection(),
		api3.WithDisplayNamePostProcessors(
			api3.SlashesToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	getObjects, err := explorer.ReadObjects("GET",
		api3.AndPathMatcher{
			api3.NewAllowPathStrategy(pathsOf(getEndpoints)),
			api3.NestedIDPathIgnorer{},
		},
		getEndpoints,
		displayNameOverride,
		nil,
	)
	goutils.MustBeNil(err)

	postObjects, err := explorer.ReadObjectsPost(
		api3.NewAllowPathStrategy(pathsOf(postEndpoints)),
		postEndpoints,
		displayNameOverride,
		nil,
	)
	goutils.MustBeNil(err)

	return append(getObjects, postObjects...)
}
