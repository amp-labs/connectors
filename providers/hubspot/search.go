package hubspot

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
)

// Search uses the POST /search endpoint to filter object records and return the result.
// This endpoint has a limit of 10,000 records. If the result has more than 10,000 records,
// the caller should employ sorting to paginate through the result on the client side.
// This endpoint paginates using paging.next.after which is to be used as an offset.
// Archived results do not appear in search results.
// Read more @ https://developers.hubspot.com/docs/api/crm/search
func (c *Connector) Search(ctx context.Context, config SearchParams) (*common.ReadResult, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if crmObjectsWithoutPropertiesAPISupport.Has(config.ObjectName) {
		// Objects outside ObjectAPI have different endpoint while both are part of CRM module.
		// For instance such object is Lists.
		return c.searchCRM(ctx, searchCRMParams{
			SearchParams: config,
		})
	}

	url, err := c.getCRMObjectsSearchURL(config)
	if err != nil {
		return nil, err
	}

	// TODO (good) validate underlying getURL
	rsp, err := c.Client.Post(ctx, url, makeFilterBody(config))
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		getNextRecordsAfter,
		c.getDataMarshaller(ctx, config.ObjectName, config.AssociatedObjects),
		config.Fields,
	)
}

// searchCRM is intended for objects outside HubSpot's ObjectAPI.
// For objects within ObjectAPI, refer to the Search method.
//
// Case-by-case explanation:
// * Lists
//   - Provider API endpoint for search
//     https://developers.hubspot.com/docs/guides/api/crm/lists/overview#search-for-a-list
//   - Search always returns an array of items, unlike the usual "read" operation.
//     Therefore, the "retrieve" API endpoint is not used
//     https://developers.hubspot.com/docs/guides/api/crm/lists/overview#retrieve-lists
func (c *Connector) searchCRM(
	ctx context.Context, config searchCRMParams,
) (*common.ReadResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getCRMSearchURL(config)
	if err != nil {
		return nil, err
	}

	payload, err := config.payload()
	if err != nil {
		return nil, err
	}

	// TODO (good) validate underlying getURL
	rsp, err := c.Client.Post(ctx, url, payload)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		common.ExtractOptionalRecordsFromPath(config.ObjectName),
		getNextRecordsURLCRM,
		common.GetMarshaledData,
		config.Fields,
	)
}

// BuildLastModifiedFilterGroup filters records modified since the given time.
// If the time is zero, it returns an empty filter. For contacts, it uses the
// lastmodifieddate field. For other objects, it uses the hs_lastmodifieddate.
// Read more: https://community.hubspot.com/t5/APIs-Integrations/CRM-V3-API-Search-issue-with-Contacts-when-using-Filters/m-p/324617
//
//nolint:lll
func BuildLastModifiedFilterGroup(params *common.ReadParams) Filter {
	if params.Since.IsZero() {
		return Filter{}
	}

	// Use the lastmodifieddate field for contacts, and hs_lastmodifieddate for other objects.
	lastModifiedField := ObjectFieldHsLastModifiedDate
	if naming.PluralityAndCaseIgnoreEqual(params.ObjectName, string(ObjectTypeContact)) {
		lastModifiedField = ObjectFieldLastModifiedDate
	}

	return Filter{
		FieldName: string(lastModifiedField),
		Operator:  FilterOperatorTypeGTE,
		Value:     params.Since.Format(time.RFC3339),
	}
}

// BuildUntilTimestampFilterGroup filters records modified until and including the given time.
func BuildUntilTimestampFilterGroup(params *common.ReadParams) Filter {
	if params.Until.IsZero() {
		return Filter{}
	}

	// Use the lastmodifieddate field for contacts, and hs_lastmodifieddate for other objects.
	lastModifiedField := ObjectFieldHsLastModifiedDate
	if params.ObjectName == string(ObjectTypeContact) {
		lastModifiedField = ObjectFieldLastModifiedDate
	}

	return Filter{
		FieldName: string(lastModifiedField),
		Operator:  FilterOperatorTypeLTE,
		Value:     params.Until.Format(time.RFC3339),
	}
}

// BuildIdFilterGroup filters records greater than the given id.
func BuildIdFilterGroup(id string) Filter {
	return Filter{
		FieldName: string(ObjectFieldHsObjectId),
		Operator:  FilterOperatorTypeGT,
		Value:     id,
	}
}

// BuildSort builds a sort by clause for the given field and direction.
func BuildSort(field ObjectField, dir SortDirection) SortBy {
	return SortBy{
		PropertyName: string(field),
		Direction:    dir,
	}
}

func makeFilterBody(config SearchParams) map[string]any {
	filterBody := map[string]any{
		"limit": DefaultPageSize,
	}

	if config.FilterGroups != nil {
		filterBody["filterGroups"] = config.FilterGroups
	}

	if config.NextPage != "" {
		filterBody["after"] = config.NextPage
	}

	if config.SortBy != nil {
		filterBody["sorts"] = config.SortBy
	}

	if config.Fields != nil {
		filterBody["properties"] = config.Fields.List()
	}

	return filterBody
}
