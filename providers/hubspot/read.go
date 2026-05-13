package hubspot

import (
	"context"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/hubspot/internal/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
)

// Read reads data from Hubspot. If Since is set, it will use the
// ReadUsingSearchAPI endpoint instead to filter records, but it will be
// limited to a maximum of 10,000 records. This is a limit of the
// search endpoint. If Since is not set, it will use the read endpoint.
// In case Deleted objects won’t appear in any search results.
// Deleted objects can only be read by using this endpoint.
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) { //nolint:funlen
	ctx = logging.With(ctx, "connector", "hubspot")

	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	switch {
	case core.CRMObjectsWithoutPropertiesAPISupport.Has(params.ObjectName):
		// Object is part of CRM namespace but outside ObjectAPI.
		// For instance object "lists" is returned only via CRM Search endpoint.
		return c.searchCRM(ctx, searchCRMParams{
			SearchParams: SearchParams{
				ObjectName: params.ObjectName,
				NextPage:   params.NextPage,
				Fields:     params.Fields,
			},
		})
	case core.MarketingObjects.Has(params.ObjectName):
		// Object is part of Hubspot Marketing API.
		return c.readMarketing(ctx, params, core.MarketingObjects[params.ObjectName])
	case core.MiscellaneousObjects.Has(params.ObjectName):
		return c.readMiscAPI(ctx, params, core.MiscellaneousObjects[params.ObjectName])
	default:
		// Otherwise object belongs to Hubspot Objects API (sub-category of CRM namespace).
		return c.readCRMObjectsAPI(ctx, params)
	}
}

func (c *Connector) readCRMObjectsAPI(
	ctx context.Context, config common.ReadParams,
) (*common.ReadResult, error) { //nolint:funlen
	// If filtering is required, then we have to use the search endpoint.
	// The Search endpoint has a 10K record limit. In case this limit is reached,
	// the sorting allows the caller to continue in another call by offsetting
	// until the ID of the last record that was successfully fetched.
	filters := make(Filters, 0)
	if !config.Since.IsZero() {
		filters = append(filters, BuildLastModifiedFilterGroup(&config))
	}

	if !config.Until.IsZero() {
		filters = append(filters, BuildUntilTimestampFilterGroup(&config))
	}

	filters = append(filters, BuildBuilderFilters(config.BuilderFilter)...)

	if len(filters) != 0 {
		searchParams := SearchParams{
			ObjectName: config.ObjectName,
			FilterGroups: []FilterGroup{{
				Filters: filters,
				// Add more filter groups to OR them together
			}},
			SortBy: []SortBy{
				BuildSort(ObjectFieldHsObjectId, SortDirectionAsc),
			},
			NextPage:          config.NextPage,
			Fields:            config.Fields,
			AssociatedObjects: config.AssociatedObjects,
		}

		return c.ReadUsingSearchAPI(ctx, searchParams)
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.JSONHTTPClient().Get(ctx, url)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		core.GetRecords,
		core.GetNextRecordsURL,
		associations.CreateDataMarshallerWithAssociations(
			ctx, c.associationsFiller, config.ObjectName, config.AssociatedObjects),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(params common.ReadParams) (string, error) {
	if len(params.NextPage) != 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		return params.NextPage.String(), nil
	}

	// If NextPage is not set, then we're reading the first page of results.
	// We need to construct the query and then make the request.
	url, err := c.getCRMObjectsURL(params.ObjectName)
	if err != nil {
		return "", err
	}

	fields := params.Fields.List()
	if len(fields) != 0 {
		url.WithQueryParam("properties", strings.Join(fields, ","))
	}

	if params.Deleted {
		url.WithQueryParam("archived", "true")
	}

	url.WithQueryParam("limit", core.DefaultPageSize)

	return url.String(), nil
}

func (c *Connector) readMarketing(ctx context.Context,
	params common.ReadParams, object core.ObjectDescription,
) (*common.ReadResult, error) {
	url, err := c.buildMarketingReadURL(params, &object)
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	identifier := "id"
	if params.ObjectName == core.ObjectMarketingEvents {
		identifier = "objectId"
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("results"),
		makeIncrementalFilterFunc(params),
		readhelper.MakeMarshaledDataFuncWithId(
			object.RecordTransformer,
			readhelper.IdFieldQuery{Field: identifier},
		),
		params.Fields,
	)
}

// When reading objects in Hubspot you must explicitly request the fields.
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/guide#campaign-properties
//
// Reading campaigns object:
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/get-campaigns
//   - Incremental reading is not available.
//   - Sorting is applied using "updatedAt" field from newest to oldest.
func (c *Connector) buildMarketingReadURL(
	params common.ReadParams, object *core.ObjectDescription,
) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getMarketingURL(object)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, core.DefaultPageSize))

	if params.ObjectName == core.ObjectMarketingForms || params.ObjectName == core.ObjectMeetingLinks {
		// This object does not have such query params. For consistency, it is reflected here.
		// Sending non-existent query params is not considered an error by provider.
	} else {
		url.WithQueryParam("properties", strings.Join(params.Fields.List(), ","))
		url.WithQueryParam("sort", "-updatedAt") // newest first
	}

	return url, nil
}

// makeIncrementalFilterFunc embodies connector-side filtering.
// ReverseOrder is used because we request Campaigns sorted from newest to oldest.
func makeIncrementalFilterFunc(params common.ReadParams) common.RecordsFilterFunc {
	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(core.GetNextRecordsURL)
	}

	order := readhelper.ReverseOrder
	if params.ObjectName == core.ObjectMarketingForms {
		order = readhelper.Unordered
	}

	return readhelper.MakeTimeFilterFunc(
		order,
		readhelper.NewTimeBoundary(),
		"updatedAt", time.RFC3339,
		core.GetNextRecordsURL,
	)
}

func (c *Connector) readMiscAPI(ctx context.Context,
	params common.ReadParams, object core.ObjectDescription,
) (*common.ReadResult, error) {
	url, err := c.buildMiscURL(params, &object)
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("results"),
		readhelper.MakeTimeFilterFunc(
			readhelper.Unordered,
			readhelper.NewTimeBoundary(),
			"updatedAt", time.RFC3339,
			core.GetNextRecordsURL,
		),
		readhelper.MakeMarshaledDataFuncWithId(
			object.RecordTransformer,
			readhelper.IdFieldQuery{Field: "id"},
		),
		params.Fields,
	)
}

func (c *Connector) buildMiscURL(
	params common.ReadParams, object *core.ObjectDescription,
) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.rootURL(object.Path)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, object.PageSize))

	return url, nil
}
