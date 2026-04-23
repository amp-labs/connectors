package salesforce

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/associations"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/metadata"
)

const defaultSOQLPageSize = 2000

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

	return common.ParseResult(
		rsp,
		core.GetRecords,
		core.GetNextRecordsURL,
		core.GetDataMarshallerForRead(config),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		return c.getDomainURL(config.NextPage.String())
	}

	// If NextPage is not set, then we're reading the first page of results.
	// We need to construct the SOQL query and then make the request.
	// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_query.htm
	url, err := c.getRestApiURL("query")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("q", makeSOQL(config, c.getTimestampColumn()).String())

	return url, nil
}

// makeSOQL returns the SOQL query for the desired read operation.
// The timestampColumn parameter specifies which field to use for Since/Until filtering
// (typically "SystemModstamp").
func makeSOQL(params common.ReadParams, timestampColumn string) *core.SOQLBuilder {
	fields := associations.FieldsForSelectQueryRead(&params)
	soql := (&core.SOQLBuilder{}).SelectFields(fields).From(params.ObjectName)
	addWhereClauses(soql, params, timestampColumn)

	return soql
}

// addWhereClauses adds WHERE clauses to the SOQL query based on the config.
func addWhereClauses(soql *core.SOQLBuilder, config common.ReadParams, timestampColumn string) {
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

	if config.Filter != "" {
		soql.Where(config.Filter)
	}

	if config.BuilderFilter != nil {
		for _, ff := range config.BuilderFilter.FieldFilters {
			if ff.Operator != common.FilterOperatorEQ {
				continue
			}

			soql.Where(buildSOQLEqCondition(ff.FieldName, ff.Value))
		}
	}

	if config.PageSize > 0 {
		soql.Limit(int64(config.PageSize))
	}
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

	triggerCodeMap := make(map[common.ObjectName]string, len(triggerParams))
	for objName, params := range triggerParams {
		triggerCodeMap[objName] = metadata.GenerateTriggerCodeForFilteredRead(*params, params.IndicatorField.FieldName)
	}

	zipDataMap, err := buildApexTriggerZips(triggerParams, triggerCodeMap)
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
