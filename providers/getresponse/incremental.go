package getresponse

import (
	"net/url"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	ResponseAtKey = "createdOn"
)

// makeFilterFunc returns a filtering function that respects the Since and Until parameters specified in ReadParams.
//
// If the object supports server-side filtering and it was applied, client-side filtering is skipped
// (since the server already filtered the results).
//
// If the object does not expose a suitable timestamp field for filtering,
// the function returns an identity filter (passing all records unchanged).
//
// Note:
//
//	Even if the HTTP request does not support server-side time filtering,
//	the client still applies local filtering to exclude records outside the specified time window.
func makeFilterFunc(params common.ReadParams, requestURL *url.URL) common.RecordsFilterFunc {
	timeSpec := objectsFilterParam.Get(params.ObjectName)

	// If server-side filtering is supported and was used, skip client-side filtering
	// (the server already filtered the results, so no need to filter again)
	if timeSpec.filterType == serverSideFilter && (!params.Since.IsZero() || !params.Until.IsZero()) {
		// Server already filtered, just handle pagination
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(requestURL))
	}

	// No timestamp field available for client-side filtering
	if timeSpec.ResponseAt == nil {
		// No timestamp field available; return all records unfiltered.
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(requestURL))
	}

	// Apply client-side filtering for objects that don't support server-side filtering
	// or when server-side filtering wasn't requested (no Since/Until params)
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
//   - filterType: whether the object supports server-side filtering via query[createdOn][from] and query[createdOn][to].
//   - ResponseAt: the timestamp field to use when advancing the incremental read.
//
// Notes:
//   - Some objects support server-side filtering via query[createdOn][from] and query[createdOn][to].
//   - Some objects don't support server-side filtering but expose a timestamp field usable for client-side filtering.
//   - If both are unavailable, filterType = clientSideFilter and ResponseAt = "".
//
// References:
//
//	GetResponse API v3 documentation: https://apireference.getresponse.com/
var objectsFilterParam = datautils.NewDefaultMap(map[string]timeFieldSpec{ // nolint:gochecknoglobals
	// -----------------------------------------------------------------------------
	// Objects with server-side filtering support (query[createdOn][from] and query[createdOn][to])
	// Based on GetResponse API v3 OpenAPI specification analysis
	// -----------------------------------------------------------------------------
	// addresses: https://apireference.getresponse.com/#operation/getAddressList
	"addresses": {filterType: serverSideFilter, ResponseAt: nil},
	// autoresponders: https://apireference.getresponse.com/#operation/getAutoresponderList
	"autoresponders": {filterType: serverSideFilter, ResponseAt: nil},
	// campaigns: https://apireference.getresponse.com/#operation/getCampaignList
	"campaigns": {filterType: serverSideFilter, ResponseAt: nil},
	// click-tracks: https://apireference.getresponse.com/#operation/getClickTrackList
	"click-tracks": {filterType: serverSideFilter, ResponseAt: nil},
	// contacts: https://apireference.getresponse.com/#operation/getContactList
	"contacts": {filterType: serverSideFilter, ResponseAt: nil},
	// forms: https://apireference.getresponse.com/#operation/getLegacyFormList
	"forms": {filterType: serverSideFilter, ResponseAt: nil},
	// imports: https://apireference.getresponse.com/#operation/getImportList
	"imports": {filterType: serverSideFilter, ResponseAt: nil},
	// landing-pages: https://apireference.getresponse.com/#operation/getLpsList
	"landing-pages": {filterType: serverSideFilter, ResponseAt: nil},
	// newsletters: https://apireference.getresponse.com/#operation/getNewsletterList
	"newsletters": {filterType: serverSideFilter, ResponseAt: nil},
	// rss-newsletters: https://apireference.getresponse.com/#operation/getRssNewslettersList
	"rss-newsletters": {filterType: serverSideFilter, ResponseAt: nil},
	// search-contacts: https://apireference.getresponse.com/#operation/newSearchContacts
	"search-contacts": {filterType: serverSideFilter, ResponseAt: nil},
	// splittests: https://apireference.getresponse.com/#operation/getSplittestList
	"splittests": {filterType: serverSideFilter, ResponseAt: nil},
	// suppressions: https://apireference.getresponse.com/#operation/getSuppressionsList
	"suppressions": {filterType: serverSideFilter, ResponseAt: nil},
	// -----------------------------------------------------------------------------
	// Objects without server-side filtering but with timestamp fields for client-side filtering
	// -----------------------------------------------------------------------------
	// custom-events: Has createdOn field but doesn't support query[createdOn] filters
	"custom-events": {filterType: clientSideFilter, ResponseAt: &ResponseAtKey},
	// files: Has createdOn field but doesn't support query[createdOn] filters
	"files": {filterType: clientSideFilter, ResponseAt: &ResponseAtKey},
	// folders: Has createdOn field but doesn't support query[createdOn] filters
	"folders": {filterType: clientSideFilter, ResponseAt: &ResponseAtKey},
	// from-fields: Has createdOn field but doesn't support query[createdOn] filters
	"from-fields": {filterType: clientSideFilter, ResponseAt: &ResponseAtKey},
	// gdpr-fields: Has createdOn field but doesn't support query[createdOn] filters
	"gdpr-fields": {filterType: clientSideFilter, ResponseAt: &ResponseAtKey},
	// templates: Has createdOn field but doesn't support query[createdOn] filters
	"templates": {filterType: clientSideFilter, ResponseAt: &ResponseAtKey},
	// webinars: Has createdOn field but doesn't support query[createdOn] filters
	"webinars": {filterType: clientSideFilter, ResponseAt: &ResponseAtKey},
}, func(objectName string) timeFieldSpec {
	// Default case: no time-based filter or response field available.
	// Unknown objects default to clientSideFilter to avoid adding unsupported query parameters.
	return timeFieldSpec{
		filterType: clientSideFilter,
		ResponseAt: nil,
	}
})

// timeFieldSpec defines which GetResponse API parameters and response fields are
// associated with time-based filtering for incremental reads.
//
// Example: request records created after a certain time, and advance incrementally based on their "createdOn" timestamp.
//
//	timeFieldSpec{
//	    filterType: clientSideFilter,
//	    ResponseAt: "createdOn",
//	}
type timeFieldSpec struct {
	// filterType specifies whether the object supports server-side filtering via query[createdOn][from] and query[createdOn][to].
	filterType filterParameterType
	// ResponseAt specifies the timestamp field available in the response used
	// to advance the incremental cursor.
	ResponseAt *string
}

type filterParameterType int

const (
	// clientSideFilter means the object does not support server-side time filtering.
	clientSideFilter filterParameterType = iota
	// serverSideFilter means the object supports server-side filtering via query[createdOn][from] and query[createdOn][to].
	serverSideFilter
)

// shouldAddServerSideFilter determines whether to add server-side since/until filters
// to the request URL based on the object's capabilities.
func shouldAddServerSideFilter(objectName string, params common.ReadParams) bool {
	timeSpec := objectsFilterParam.Get(objectName)
	return timeSpec.filterType == serverSideFilter && (!params.Since.IsZero() || !params.Until.IsZero())
}
