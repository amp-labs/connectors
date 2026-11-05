package meeting

import (
	"strings"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/filters"
	mapping "github.com/amp-labs/connectors/scripts/openapi/internal/api/map"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/reducers"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/scripts/openapi/internal/core"
	"github.com/amp-labs/connectors/scripts/openapi/zoom/internal/files"
)

const (
	tagSelfUser = "self-user"
	tagAllUsers = "all-users"
)

var (
	allowedEndpoints = filters.NewRules( // nolint:gochecknoglobals
		"/archive_files",
		"/devices/groups",
		"/devices",
		"/meetings/meeting_summaries",
		"/report/billing",
		"/report/activities",
		"/sip_phones/phones",
		"/tracking_fields",
		"/h323/devices",
		"/users/{userId}/recordings",
		"/users/{userId}/meetings",
		"/users/{userId}/upcoming_meetings",
		"/report/users/{userId}/meetings",
		"/users/{userId}/meeting_templates",
		"/users/{userId}/tsp",
		"/users/{userId}/webinars",
		"/report/users",
		"/report/cloud_recording",
		"/report/operationlogs",
		"/report/meeting_activities",
		"/report/telephone",
		"/report/upcoming_events",
	)

	objectEndpoints = core.NewTrackedMap(map[string]string{ // nolint:gochecknoglobals
		"/devices/groups":                 "devices_groups",
		"/meetings/meeting_summaries":     "meeting_summaries",
		"/report/billing":                 "billing_report",
		"/report/activities":              "activities_report",
		"/h323/devices":                   "h323_devices",
		"/report/users/{userId}/meetings": "meetings_report",
		"/report/users":                   "users_report",
		"/report/cloud_recording":         "recordings_report",
		"/report/operationlogs":           "operation_logs_report",
		"/report/meeting_activities":      "meeting_activities_report",
		"/report/telephone":               "telephone_report",
		"/report/upcoming_events":         "upcoming_events_report",
	})

	displayNameOverride = core.NewTrackedMap(map[string]string{ // nolint:gochecknoglobals
		"/h323/devices": "H.323/SIP Devices",
	})
)

func Objects() pipeline.Pipeline[spec.Schema] {
	extractor, err := files.MeetingFile.Extractor()
	goutils.MustBeNil(err)

	getSchemas, err := extractor.ExtractListSchemas(api.GET,
		api.List.WithAutoSelectArrayItems(),
	)
	goutils.MustBeNil(err)

	pipe := pipeline.NewSchemaPipe(getSchemas).
		Filter(filters.KeepWithPath(
			filters.NewAllowPathStrategy(allowedEndpoints),
		)).
		Map(mapping.ObjectNameOverride(objectEndpoints)).
		Reduce(reducers.ShortestNameFromUrl).
		Reduce(endpointsWithRequiredUserId).
		Map(mapping.DisplayNameFromObjectName).
		Map(mapping.DisplayNameFormat(
			mapping.CamelCaseToSpaceSeparated,
			mapping.CapitalizeFirstLetterEveryWord,
		)).
		Map(mapping.DisplayNameOverride(displayNameOverride)).
		Map(displayNameForUserBasedSchema)

	output.PrintUnusedRules("meeting/allowedEndpoints", allowedEndpoints)
	output.PrintRedundantKeys("meeting/objectEndpoints", objectEndpoints)
	output.PrintRedundantKeys("meeting/displayNameOverride", displayNameOverride)

	return pipe
}

func endpointsWithRequiredUserId(schemas []spec.Schema) []spec.Schema {
	result := make([]spec.Schema, 0, len(schemas))
	for _, schema := range schemas {
		if strings.Contains(schema.UrlPath, "{userId}") {
			schemaForAccounts := schema.Clone()
			schemaForSelf := schema.Clone()

			schemaForSelf.UrlPath = strings.ReplaceAll(schemaForSelf.UrlPath, "{userId}", "me")
			schemaForSelf.Tags.AddOne(tagSelfUser)

			schemaForAccounts.ObjectName = "account_" + schemaForAccounts.ObjectName
			schemaForAccounts.Tags.AddOne(tagAllUsers)

			result = append(result, schemaForAccounts, schemaForSelf)
		} else {
			result = append(result, schema)
		}
	}

	return result
}

func displayNameForUserBasedSchema(schema spec.Schema) spec.Schema {
	if schema.Tags.Has(tagSelfUser) {
		// Ex: "My Meetings"
		schema.DisplayName = "My " + schema.DisplayName
	}

	if schema.Tags.Has(tagAllUsers) {
		// Ex: "Account Meetings Report" becomes "All Meetings Report"
		name, _ := strings.CutPrefix(schema.DisplayName, "Account ")
		schema.DisplayName = "All " + name
	}

	return schema
}
