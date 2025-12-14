package getresponse

import (
	"net/url"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	ResponseAtKey = "createdOn" // nolint:gochecknoglobals
)

// makeFilterFunc returns a filtering function that respects the Since and Until parameters specified in ReadParams.
//
// If the object supports provider-side filtering and it was applied, connector-side filtering is skipped
// (since the provider already filtered the results).
//
// If the object does not expose a suitable timestamp field for filtering,
// the function returns an identity filter (passing all records unchanged).
//
// Note:
//
//	Even if the HTTP request does not support provider-side time filtering,
//	the connector still applies local filtering to exclude records outside the specified time window.
func makeFilterFunc(params common.ReadParams, requestURL *url.URL) common.RecordsFilterFunc {
	timeSpec := objectsFilterParam.Get(params.ObjectName)

	// If provider-side filtering is supported and was used, skip connector-side filtering
	// (the provider already filtered the results, so no need to filter again)
	if timeSpec.filterType == providerSideFilter && (!params.Since.IsZero() || !params.Until.IsZero()) {
		// Provider already filtered, just handle pagination
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(requestURL))
	}

	// No timestamp field available for connector-side filtering
	if timeSpec.ResponseAt == nil {
		// No timestamp field available; return all records unfiltered.
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(requestURL))
	}

	// Apply connector-side filtering for objects that don't support provider-side filtering
	// or when provider-side filtering wasn't requested (no Since/Until params)
	return readhelper.MakeTimeFilterFunc(
		readhelper.ChronologicalOrder,
		readhelper.NewTimeBoundary(),
		*timeSpec.ResponseAt, time.RFC3339,
		makeNextRecordsURL(requestURL))
}

// objectsFilterParam defines which timestamp fields and filters apply to
// each GetResponse object for incremental reads.
//
// Structure:
//   - filterType: whether the object supports provider-side filtering via
//     query[createdOn][from] and query[createdOn][to].
//   - ResponseAt: the timestamp field to use when advancing the incremental read.
//
// Notes:
//   - Some objects support provider-side filtering via query[createdOn][from] and query[createdOn][to].
//   - Some objects don't support provider-side filtering but expose a timestamp field usable for connector-side filtering.
//   - If both are unavailable, filterType = connectorSideFilter and ResponseAt = "".
//
// References:
//
//	GetResponse API v3 documentation: https://apireference.getresponse.com/
var objectsFilterParam = datautils.NewDefaultMap(map[string]timeFieldSpec{ // nolint:gochecknoglobals
	// -----------------------------------------------------------------------------
	// Objects with provider-side filtering support (query[createdOn][from] and query[createdOn][to])
	// Based on GetResponse API v3 OpenAPI specification analysis
	// -----------------------------------------------------------------------------
	// addresses: https://apireference.getresponse.com/#operation/getAddressList
	"addresses": {filterType: providerSideFilter, ResponseAt: nil},
	// autoresponders: https://apireference.getresponse.com/#operation/getAutoresponderList
	"autoresponders": {filterType: providerSideFilter, ResponseAt: nil},
	// campaigns: https://apireference.getresponse.com/#operation/getCampaignList
	"campaigns": {filterType: providerSideFilter, ResponseAt: nil},
	// click-tracks: https://apireference.getresponse.com/#operation/getClickTrackList
	"click-tracks": {filterType: providerSideFilter, ResponseAt: nil},
	// contacts: https://apireference.getresponse.com/#operation/getContactList
	"contacts": {filterType: providerSideFilter, ResponseAt: nil},
	// forms: https://apireference.getresponse.com/#operation/getLegacyFormList
	"forms": {filterType: providerSideFilter, ResponseAt: nil},
	// imports: https://apireference.getresponse.com/#operation/getImportList
	"imports": {filterType: providerSideFilter, ResponseAt: nil},
	// landing-pages: https://apireference.getresponse.com/#operation/getLpsList
	"landing-pages": {filterType: providerSideFilter, ResponseAt: nil},
	// newsletters: https://apireference.getresponse.com/#operation/getNewsletterList
	"newsletters": {filterType: providerSideFilter, ResponseAt: nil},
	// rss-newsletters: https://apireference.getresponse.com/#operation/getRssNewslettersList
	"rss-newsletters": {filterType: providerSideFilter, ResponseAt: nil},
	// search-contacts: https://apireference.getresponse.com/#operation/newSearchContacts
	"search-contacts": {filterType: providerSideFilter, ResponseAt: nil},
	// splittests: https://apireference.getresponse.com/#operation/getSplittestList
	"splittests": {filterType: providerSideFilter, ResponseAt: nil},
	// suppressions: https://apireference.getresponse.com/#operation/getSuppressionsList
	"suppressions": {filterType: providerSideFilter, ResponseAt: nil},
	// -----------------------------------------------------------------------------
	// Objects without provider-side filtering but with timestamp fields for connector-side filtering
	// -----------------------------------------------------------------------------
	// custom-events: Has createdOn field but doesn't support query[createdOn] filters
	"custom-events": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// files: Has createdOn field but doesn't support query[createdOn] filters
	"files": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// folders: Has createdOn field but doesn't support query[createdOn] filters
	"folders": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// from-fields: Has createdOn field but doesn't support query[createdOn] filters
	"from-fields": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// gdpr-fields: Has createdOn field but doesn't support query[createdOn] filters
	"gdpr-fields": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// templates: Has createdOn field but doesn't support query[createdOn] filters
	"templates": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// webinars: Has createdOn field but doesn't support query[createdOn] filters
	"webinars": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
}, func(objectName string) timeFieldSpec {
	// Default case: no time-based filter or response field available.
	// Unknown objects default to connectorSideFilter to avoid adding unsupported query parameters.
	return timeFieldSpec{
		filterType: connectorSideFilter,
		ResponseAt: nil,
	}
})

// timeFieldSpec defines which GetResponse API parameters and response fields are
// associated with time-based filtering for incremental reads.
//
// Example: request records created after a certain time, and advance incrementally
// based on their "createdOn" timestamp.
//
//	timeFieldSpec{
//	    filterType: connectorSideFilter,
//	    ResponseAt: "createdOn",
//	}
type timeFieldSpec struct {
	// filterType specifies whether the object supports provider-side filtering via
	// query[createdOn][from] and query[createdOn][to].
	filterType filterParameterType
	// ResponseAt specifies the timestamp field available in the response used
	// to advance the incremental cursor.
	ResponseAt *string
}

type filterParameterType int

const (
	// connectorSideFilter means the object does not support provider-side time filtering.
	connectorSideFilter filterParameterType = iota
	// providerSideFilter means the object supports provider-side filtering via
	// query[createdOn][from] and query[createdOn][to].
	providerSideFilter
)

// shouldAddProviderSideFilter determines whether to add provider-side since/until filters
// to the request URL based on the object's capabilities.
func shouldAddProviderSideFilter(objectName string, params common.ReadParams) bool {
	timeSpec := objectsFilterParam.Get(objectName)
	return timeSpec.filterType == providerSideFilter && (!params.Since.IsZero() || !params.Until.IsZero())
}
