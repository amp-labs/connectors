// Extracts list endpoint schemas from OpenAPI spec and writes providers/mailgun/metadata/schemas.json.
package main

import (
	"log"
	"log/slog"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
	"github.com/amp-labs/connectors/providers/mailgun/openapi"
	utilsopenapi "github.com/amp-labs/connectors/scripts/openapi/utils"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

//
//nolint:gochecknoglobals
var ignoreEndpoints = []string{
	// singular objects
	"/v3/ips/account/settings",
	"/v3/ips/request/new",
	"/v1/analytics/tags/limits",
	"/v5/accounts/http_signing_key",
	"/v5/accounts/limit/custom/monthly",
	"/v3/routes/match",
	"/v5/users/me",
	// array of strings not objects
	"/v1/alerts/events",
	// map / free-form responses
	"/v1/bounce-classification/config/entities",
	"/v1/bounce-classification/config/rules",
	// aggregate / stats responses, not list resources
	"/v3/stats/total",
	"/v3/stats/total/domains",
	"/v3/stats/filter",
}

// Nested Mailgun list routes where the last path segment is not a good object name.
//
//nolint:gochecknoglobals
var objectEndpoints = map[string]string{
	"/v5/accounts/subaccounts/ip_pools/all": "accounts/subaccounts/ip_pools",
	"/v5/accounts/subaccounts":              "accounts/subaccounts",
	"/v3/domains/dynamic_pools/assignable":  "domains/dynamic_pools/assignable",
	"/v3/dynamic_pools":                     "dynamic_pools",
	"/v3/ips/details/all":                   "ips/details",
	"/v1/dynamic_pools/domains":             "dynamic_pools/domains",
	"/v1/dynamic_pools/history":             "dynamic_pools/history",
	"/v1/thresholds/alerts/send":            "thresholds/alerts/send",
	"/v1/thresholds/limits":                 "thresholds/limits",
	"/v1/thresholds/hits":                   "thresholds/hits",
	"/v1/alerts/settings":                   "alerts/settings",
	"/v1/alerts/slack/channels":             "alerts/slack/channels",
	"/v3/lists/pages":                       "lists/pages",
	"/v5/sandbox/auth_recipients":           "sandbox/auth_recipients",
	"/v1/dkim/keys":                         "dkim/keys",
	"/v1/bounce-classification/stats":       "bounce-classification/stats",
	"/v1/bounce-classification/domains":     "bounce-classification/domains",
}

// Display names taken from OpenAPI tags/summaries in openapi/schema.yaml.
// Prefer the operation tag when it names the resource; otherwise use the summary
// with leading list/get verbs removed.
//
//nolint:gochecknoglobals
var displayNameOverrides = map[string]string{
	// tag: Delegated DIPPs
	"accounts/subaccounts/ip_pools": "Delegated DIPPs",
	// tag: Subaccounts
	"accounts/subaccounts": "Subaccounts",
	// summary: List assignable domains
	"domains/dynamic_pools/assignable": "Assignable Domains",
	// tag: Dynamic IP Pools
	"dynamic_pools": "Dynamic IP Pools",
	// tag: IPs
	"ips": "IPs",
	// summary: List account IPs - detailed view
	"ips/details": "Account IPs - Detailed View",
	// summary: List all domains assigned to dynamic IP pools
	"dynamic_pools/domains": "Domains Assigned To Dynamic IP Pools",
	// summary: List account history
	"dynamic_pools/history": "Account History",
	// tag: Send Alerts
	"thresholds/alerts/send": "Send Alerts",
	// tag: Limits
	"thresholds/limits": "Limits",
	// summary: List account hits
	"thresholds/hits": "Account Hits",
	// summary: List Alerts
	"alerts/settings": "Alerts",
	// summary: List Slack channels
	"alerts/slack/channels": "Slack Channels",
	// tag: Mailing Lists
	"lists/pages": "Mailing Lists",
	// summary: Get authorized email recipients for a sandbox domain
	"sandbox/auth_recipients": "Authorized Email Recipients",
	// tag: Domain Keys
	"dkim/keys": "Domain Keys",
	// summary: List statistics, ordered by total bounces
	"bounce-classification/stats": "Statistics",
	// summary: List domains statistic per account
	"bounce-classification/domains": "Domains Statistic",
	// tag: IP Pools
	"ip_pools": "IP Pools",
	// tag: IP Address Warmup
	"ip_warmups": "IP Address Warmup",
	// tag: IP Allowlist
	"ip_whitelist": "IP Allowlist",
}

// Most Mailgun list responses are single-array and auto-selected.
// dynamic_pools also returns reputation IP string arrays.
func mailgunResponseLocator(objectName, fieldName string) bool {
	if objectName == "dynamic_pools" {
		return fieldName == "pools"
	}

	return api3.DataObjectLocator(objectName, fieldName)
}

func main() {
	explorer, err := openapi.FileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
			api3.Pluralize,
		),
		api3.WithArrayItemAutoSelection(),
	)
	goutils.MustBeNil(err)

	readObjects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		objectEndpoints, displayNameOverrides, mailgunResponseLocator,
	)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	registry := datautils.NamedLists[string]{}

	for _, object := range readObjects {
		if object.Problem != nil {
			slog.Error("schema not extracted",
				"objectName", object.ObjectName,
				"error", object.Problem,
			)

			continue
		}

		for _, field := range object.Fields {
			schemas.Add("", object.ObjectName, object.DisplayName, object.URLPath, object.ResponseKey,
				utilsopenapi.ConvertMetadataFieldToFieldMetadataMapV2(field), nil, object.Custom)
		}

		for _, queryParam := range object.QueryParams {
			registry.Add(queryParam, object.ObjectName)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))
	goutils.MustBeNil(metadata.FileManager.SaveQueryParamStats(scrapper.CalculateQueryParamStats(registry)))

	log.Println("Completed.")
}
