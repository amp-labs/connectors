package getresponse

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

var ResponseAtKey = "createdOn" // nolint:gochecknoglobals

// GetResponse returns timestamps in the format "2006-01-02T15:04:05+0000" (no colon in timezone offset),
// which does not comply with time.RFC3339 ("2006-01-02T15:04:05Z07:00").
// The "Z0700" layout accepts both "+0000" and "Z" suffixes.
const responseTimestampFormat = "2006-01-02T15:04:05Z0700"

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
	// No time bounds requested — skip all filtering.
	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(requestURL))
	}

	timeSpec := objectsFilterParam.Get(params.ObjectName)

	// Provider already filtered the results — no need to re-filter connector-side.
	if timeSpec.filterType == providerSideFilter {
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(requestURL))
	}

	// No timestamp field available for connector-side filtering.
	if timeSpec.ResponseAt == nil {
		return readhelper.MakeIdentityFilterFunc(makeNextRecordsURL(requestURL))
	}

	// Apply connector-side filtering using the object's timestamp field.
	return readhelper.MakeTimeFilterFunc(
		readhelper.ChronologicalOrder,
		readhelper.NewTimeBoundary(),
		*timeSpec.ResponseAt, responseTimestampFormat,
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
//   - Some objects don't support provider-side filtering but expose a timestamp field
//     usable for connector-side filtering.
//   - If both are unavailable, filterType = connectorSideFilter and ResponseAt = "".
//
// References:
//
//	GetResponse API v3 documentation: https://apireference.getresponse.com/
var objectsFilterParam = datautils.NewDefaultMap(map[string]timeFieldSpec{ // nolint:gochecknoglobals
	// -----------------------------------------------------------------------------
	// Objects with confirmed provider-side filtering support (query[createdOn][from] and query[createdOn][to])
	// Verified against live GetResponse API v3 - do NOT add objects here without live testing.
	// -----------------------------------------------------------------------------
	// contacts: https://apireference.getresponse.com/#operation/getContactList
	"contacts": {filterType: providerSideFilter, ResponseAt: nil},
	// -----------------------------------------------------------------------------
	// Objects without provider-side filtering but with createdOn field for connector-side filtering
	// -----------------------------------------------------------------------------
	// addresses: https://apireference.getresponse.com/#operation/getAddressList
	"addresses": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// autoresponders: https://apireference.getresponse.com/#operation/getAutoresponderList
	"autoresponders": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// campaigns: https://apireference.getresponse.com/#operation/getCampaignList
	// Note: API rejects query[createdOn][from] with "Not allowed search field" (confirmed via live testing)
	"campaigns": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// click-tracks: https://apireference.getresponse.com/#operation/getClickTrackList
	"click-tracks": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// custom-events: https://apireference.getresponse.com/#operation/getCustomEventsList
	"custom-events": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// files: https://apireference.getresponse.com/#operation/getFileList
	"files": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// folders: https://apireference.getresponse.com/#operation/getFolderList
	"folders": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// See https://apireference.getresponse.com/#operation/getFormList.
	"forms": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// from-fields: https://apireference.getresponse.com/#operation/getFromFieldList
	"from-fields": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// gdpr-fields: https://apireference.getresponse.com/#operation/getGdprFieldList
	"gdpr-fields": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// imports: https://apireference.getresponse.com/#operation/getImportList
	"imports": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// landing-pages: https://apireference.getresponse.com/#operation/getLpsList
	"landing-pages": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// newsletters: https://apireference.getresponse.com/#operation/getNewsletterList
	"newsletters": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// rss-newsletters: https://apireference.getresponse.com/#operation/getRssNewslettersList
	"rss-newsletters": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// search-contacts: https://apireference.getresponse.com/#operation/newSearchContacts
	"search-contacts": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// splittests: https://apireference.getresponse.com/#operation/getSplittestList
	"splittests": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// suppressions: https://apireference.getresponse.com/#operation/getSuppressionsList
	"suppressions": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// templates: https://apireference.getresponse.com/#operation/getTemplateList
	"templates": {filterType: connectorSideFilter, ResponseAt: &ResponseAtKey},
	// webinars: https://apireference.getresponse.com/#operation/getWebinarsList
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
