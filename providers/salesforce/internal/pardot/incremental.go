package pardot

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func makeFilterFunc(params common.ReadParams) common.RecordsFilterFunc {
	responseAt, found := objectsFilterParam[params.ObjectName]
	if !found {
		// No timestamp field available; return all records unfiltered.
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	// Connector side filtering.
	return readhelper.MakeTimeFilterFunc(
		readhelper.ReverseOrder,
		readhelper.NewTimeBoundary(),
		responseAt, time.RFC3339,
		nextPageFunc,
	)
}

func nextPageFunc(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextPageUrl", "")
}

// Used for connector side filtering.
var objectsFilterParam = map[string]string{ // nolint:gochecknoglobals
	// Objects that have "updatedAt" field:
	"custom-redirects":           "updatedAt",
	"engagement-studio-programs": "updatedAt",
	"files":                      "updatedAt",
	"folder-contents":            "updatedAt",
	"folders":                    "updatedAt",
	"forms":                      "updatedAt",
	"layout-templates":           "updatedAt",
	"prospect-accounts":          "updatedAt",
	"tracker-domains":            "updatedAt",
	"visitors":                   "updatedAt",
}

// Used for provider side filtering.
var incrementalSinceQuery = map[string]string{ // nolint:gochecknoglobals
	// https://developer.salesforce.com/docs/marketing/pardot/guide/campaign-v5.html#query
	"campaigns": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/dynamic-content-v5.html#dynamic-content-query
	"dynamic-contents": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/email-v5.html
	"emails": "sentAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/email-template-v5.html#email-template-query
	"email-templates": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/export-v5.html
	"exports": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/external-activities-v5.html
	"external-activities": "createdAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/form-field-v5.html
	"form-fields": "createdAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/form-handler-v5.html
	"form-handlers": "createdAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/form-handler-field-v5.html
	"form-handler-fields": "createdAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/import-v5.html
	"imports": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/landing-page-v5.html
	"landing-pages": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/lifecycle-stage-v5.html
	"lifecycle-stages": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/lifecycle-history-v5.html
	"lifecycle-histories": "createdAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/list-v5.html
	"lists": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/list-email-v5.html
	"list-emails": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/list-membership-v5.html
	"list-memberships": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/opportunity-v5.html
	"opportunities": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/prospect-v5.html
	"prospects": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/tag-v5.html
	"tags": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/tagged-object-v5.html
	"tagged-objects": "createdAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/user-v5.html
	"users": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/visitor-activity-v5.html
	"visitor-activities": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/visit-v5.html
	"visits": "updatedAtAfterOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/visitor-page-view-v5.html
	"visitor-page-views": "createdAtAfterOrEqualTo",
}

// Used for provider side filtering.
var incrementalUntilQuery = map[string]string{ // nolint:gochecknoglobals
	// https://developer.salesforce.com/docs/marketing/pardot/guide/campaign-v5.html#query
	"campaigns": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/dynamic-content-v5.html#dynamic-content-query
	"dynamic-contents": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/email-v5.html
	"emails": "sentAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/email-template-v5.html#email-template-query
	"email-templates": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/export-v5.html
	"exports": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/external-activities-v5.html
	"external-activities": "createdAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/form-field-v5.html
	"form-fields": "createdAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/form-handler-v5.html
	"form-handlers": "createdAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/form-handler-field-v5.html
	"form-handler-fields": "createdAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/import-v5.html
	"imports": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/landing-page-v5.html
	"landing-pages": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/lifecycle-stage-v5.html
	"lifecycle-stages": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/lifecycle-history-v5.html
	"lifecycle-histories": "createdAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/list-v5.html
	"lists": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/list-email-v5.html
	"list-emails": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/list-membership-v5.html
	"list-memberships": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/opportunity-v5.html
	"opportunities": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/prospect-v5.html
	"prospects": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/tag-v5.html
	"tags": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/tagged-object-v5.html
	"tagged-objects": "createdAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/user-v5.html
	"users": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/visitor-activity-v5.html
	"visitor-activities": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/visit-v5.html
	"visits": "updatedAtBeforeOrEqualTo",
	// https://developer.salesforce.com/docs/marketing/pardot/guide/visitor-page-view-v5.html
	"visitor-page-views": "createdAtBeforeOrEqualTo",
}
