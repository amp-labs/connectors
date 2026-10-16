package salesforce

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/associations"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

const (
	defaultSOQLPageSize = 2000

	// afterIDPageTokenPrefix starts a NextPage token that resumes a capped read after a record Id.
	// Salesforce's own next page tokens are URL paths, so they never start with it.
	afterIDPageTokenPrefix = "after-id:"
)

var (
	errMissingRecordID = errors.New("salesforce record has no Id")

	// salesforceIDPattern matches a 15 or 18 character Salesforce record Id.
	salesforceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9]{15}(?:[a-zA-Z0-9]{3})?$`)
)

// Read reads data from Salesforce. By default, it will read all rows (backfill). However, if Since is set,
// it will read only rows that have been updated since the specified time.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if c.isPardotModule() {
		return c.pardotAdapter.Read(ctx, config)
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	result, err := common.ParseResult(
		rsp,
		core.GetRecords,
		core.GetNextRecordsURL,
		core.GetDataMarshallerForRead(config),
		config.Fields,
	)
	if err != nil {
		return nil, err
	}

	return continueCappedRead(config, rsp, result)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	afterID, err := parseAfterIDPageToken(config.NextPage)
	if err != nil {
		return nil, err
	}

	if len(config.NextPage) != 0 && afterID == "" {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		return c.getDomainURL(config.NextPage.String())
	}

	// Otherwise we're reading the first page of a query: either the first page of the read,
	// or the first page after a capped query ran out (see continueCappedRead).
	// We need to construct the SOQL query and then make the request.
	// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_query.htm
	url, err := c.getRestApiURL(core.QueryEndpoint(config.RawFilter != nil))
	if err != nil {
		return nil, err
	}

	soql, err := makeSOQL(config, c.GetTimestampColumn(common.ObjectName(config.ObjectName)))
	if err != nil {
		return nil, err
	}

	addPageSizeCap(soql, config.PageSize, afterID)

	url.WithQueryParam("q", soql.String())

	return url, nil
}

// addPageSizeCap caps the query at pageSize rows, sorted by Id, starting after afterID when it is set.
// A SOQL LIMIT caps the whole query, not one batch of it, so a capped query ends the read early
// unless the read resumes after the last Id it returned (see continueCappedRead).
func addPageSizeCap(soql *core.SOQLBuilder, pageSize int, afterID string) {
	if afterID != "" {
		soql.Where("Id > '" + afterID + "'")
	}

	if pageSize <= 0 && afterID == "" {
		return
	}

	soql.OrderBy("Id")

	if pageSize > 0 {
		soql.Limit(int64(pageSize))
	}
}

// continueCappedRead keeps a read with a PageSize going once its capped query runs out.
// Salesforce reports such a query as done after PageSize rows even when more rows match,
// so a query that returned all PageSize rows is followed by one for the rows after its last Id.
func continueCappedRead(
	config common.ReadParams,
	rsp *common.JSONHTTPResponse,
	result *common.ReadResult,
) (*common.ReadResult, error) {
	if config.PageSize <= 0 || !result.Done || len(result.Data) == 0 {
		return result, nil
	}

	body, ok := rsp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	// totalSize counts the rows of the whole query, across all of its batches.
	// If it is missing, assume the cap was reached: one extra query is cheaper than lost rows.
	totalSize, err := jsonquery.New(body).IntegerWithDefault("totalSize", int64(config.PageSize))
	if err != nil {
		return nil, err
	}

	if totalSize < int64(config.PageSize) {
		return result, nil
	}

	lastID := result.Data[len(result.Data)-1].Id
	if lastID == "" {
		return nil, fmt.Errorf("%w: cannot resume a capped read without the last record's Id", errMissingRecordID)
	}

	result.NextPage = common.NextPageToken(afterIDPageTokenPrefix + lastID)
	result.Done = false

	return result, nil
}

// parseAfterIDPageToken returns the Id a capped read resumes after,
// or an empty string when the token is not one made by continueCappedRead.
func parseAfterIDPageToken(token common.NextPageToken) (string, error) {
	afterID, ok := strings.CutPrefix(token.String(), afterIDPageTokenPrefix)
	if !ok {
		return "", nil
	}

	// The Id is placed in the SOQL as-is, so only accept a well-formed Salesforce Id.
	if !salesforceIDPattern.MatchString(afterID) {
		return "", fmt.Errorf("%w: %q", common.ErrNextPageInvalid, token)
	}

	return afterID, nil
}

// makeSOQL returns the SOQL query for the desired read operation.
// The timestampColumn parameter specifies which field to use for Since/Until filtering
// (typically "SystemModstamp").
// It fails if the params carry a raw filter that isn't a valid Salesforce raw filter.
func makeSOQL(params common.ReadParams, timestampColumn string) (*core.SOQLBuilder, error) {
	fields := associations.FieldsForSelectQueryRead(&params)
	soql := (&core.SOQLBuilder{}).SelectFields(fields).From(params.ObjectName)

	if err := addWhereClauses(soql, params, timestampColumn); err != nil {
		return nil, err
	}

	return soql, nil
}

// addWhereClauses adds WHERE clauses to the SOQL query based on the config.
func addWhereClauses(soql *core.SOQLBuilder, config common.ReadParams, timestampColumn string) error {
	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	if !config.Since.IsZero() {
		soql.Where(timestampColumn + " > " + datautils.Time.FormatRFC3339inUTC(config.Since))
	}

	if !config.Until.IsZero() {
		soql.Where(timestampColumn + " <= " + datautils.Time.FormatRFC3339inUTC(config.Until))
	}

	if config.Deleted {
		soql.Where("IsDeleted = true")
	}

	if config.RawFilter != nil {
		filter, err := common.RawFilterAs[core.SOQLFilter](*config.RawFilter)
		if err != nil {
			return err
		}

		soql.WhereRawFilter(filter.Condition)
	}

	if config.BuilderFilter != nil {
		for _, ff := range config.BuilderFilter.FieldFilters {
			if ff.Operator != common.FilterOperatorEQ {
				continue
			}

			soql.Where(buildSOQLEqCondition(ff.FieldName, ff.Value))
		}
	}

	return nil
}

// DeployApexTriggersForFilteredRead builds and deploys filtered-read apex triggers
// (datetime indicator) for the given trigger params. Each trigger sets a timestamp
// field to System.now() when any watched field changes.
func (c *Connector) DeployApexTriggersForFilteredRead( //nolint:unused
	ctx context.Context,
	triggerParams map[common.ObjectName]*ApexTriggerParams,
) (*DeployApexTriggersResult, error) {
	if len(triggerParams) == 0 {
		return &DeployApexTriggersResult{
			Results: make(map[common.ObjectName]*ApexTriggerResult),
			Errors:  make(map[common.ObjectName]error),
		}, nil
	}

	zipDataMap, err := buildApexTriggerZips(ctx, triggerParams)
	if err != nil {
		return nil, err
	}

	return c.deployApexTriggers(ctx, triggerParams, zipDataMap)
}

func (c *Connector) DefaultPageSize() int {
	return defaultSOQLPageSize
}

// buildSOQLEqCondition builds a SOQL equality condition for a field and value.
// String values are single-quoted and escaped to prevent SOQL injection.
func buildSOQLEqCondition(fieldName string, value any) string {
	switch v := value.(type) {
	case string:
		escaped := strings.ReplaceAll(v, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, "'", `\'`)

		return fmt.Sprintf("%s = '%s'", fieldName, escaped)
	case bool:
		return fmt.Sprintf("%s = %t", fieldName, v)
	default:
		return fmt.Sprintf("%s = %v", fieldName, v)
	}
}
