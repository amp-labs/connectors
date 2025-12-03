package salesforce

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

const defaultSOQLPageSize = 2000

// This helps us identify if we can use a SOQL subquery to get an associated object, since SOQL subqueries
// only work for child objects.
func getParentFieldMap() map[string]map[string]string {
	return map[string]map[string]string{
		"opportunity": {
			"account": "AccountId",
		},
	}
}

func isParentRelationship(objectName, associatedObject string) bool {
	parentFieldMap := getParentFieldMap()

	objMap, ok := parentFieldMap[strings.ToLower(objectName)]
	if !ok {
		return false
	}

	_, ok = objMap[strings.ToLower(associatedObject)]

	return ok
}

func getParentFieldName(objectName, associatedObject string) string {
	parentFieldMap := getParentFieldMap()

	objMap, ok := parentFieldMap[strings.ToLower(objectName)]
	if !ok {
		return ""
	}

	return objMap[strings.ToLower(associatedObject)]
}

// containsField checks if a field exists in the fields list (case-insensitive).
// e.g. containsField(["Id", "Name", "AccountId"], "accountid") -> true.
func containsField(fields []string, fieldName string) bool {
	fieldLower := strings.ToLower(fieldName)
	for _, field := range fields {
		if strings.ToLower(field) == fieldLower {
			return true
		}
	}

	return false
}

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
		getRecords,
		getNextRecordsURL,
		getSalesforceDataMarshaller(config),
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
	url, err := c.getRestApiURL("query")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("q", makeSOQL(config).String())

	return url, nil
}

// makeSOQL returns the SOQL query for the desired read operation.
func makeSOQL(config common.ReadParams) *core.SOQLBuilder {
	fields := addAssociationFields(config)
	soql := (&core.SOQLBuilder{}).SelectFields(fields).From(config.ObjectName)
	addWhereClauses(soql, config)

	return soql
}

// addAssociationFields adds fields for associated objects to the fields list.
func addAssociationFields(config common.ReadParams) []string {
	fields := config.Fields.List()

	if config.AssociatedObjects == nil {
		return fields
	}

	for _, obj := range config.AssociatedObjects {
		fields = addFieldForAssociation(fields, config.ObjectName, obj)
	}

	return fields
}

// addFieldForAssociation adds a field or subquery for an associated object.
func addFieldForAssociation(fields []string, objectName, assocObj string) []string {
	// Some objects cannot be queried using a subquery, such as when the associated object is a parent object.
	// In that case, we fetch the associated object's ID as a field, and fetch the full object in the q
	if isParentRelationship(objectName, assocObj) {
		parentField := getParentFieldName(objectName, assocObj)
		if parentField != "" && !containsField(fields, parentField) {
			fields = append(fields, parentField)
		}
	} else {
		// Generates subqueries like: (SELECT FIELDS(STANDARD) FROM Contacts)
		// Just standard fields for now, because salesforce errors out > 200 fields on an object.
		// Source: https://www.infallibletechie.com/2023/04/parent-child-records-in-salesforce-soql-using-rest-api.html
		fields = append(fields, "(SELECT FIELDS(STANDARD) FROM "+assocObj+")")
	}

	return fields
}

// addWhereClauses adds WHERE clauses to the SOQL query based on the config.
func addWhereClauses(soql *core.SOQLBuilder, config common.ReadParams) {
	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	if !config.Since.IsZero() {
		soql.Where("SystemModstamp > " + datautils.Time.FormatRFC3339inUTC(config.Since))
	}

	if !config.Until.IsZero() {
		soql.Where("SystemModstamp <= " + datautils.Time.FormatRFC3339inUTC(config.Until))
	}

	if config.Deleted {
		soql.Where("IsDeleted = true")
	}

	// TODO: When we support builder facing filters, we should escape the
	// filter string to avoid SOQL injection.
	if config.Filter != "" {
		soql.Where(config.Filter)
	}

	if config.PageSize > 0 {
		soql.Limit(config.PageSize)
	}
}

func (c *Connector) DefaultPageSize() int {
	return defaultSOQLPageSize
}
