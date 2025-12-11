package zoho

import (
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
)

const (
	timeLayout = "2006-01-02T15:04:05.000Z"
)

type (
	// objectNameTransformer takes an object and transfoms it to a standard zoho provider api name.
	objectNameTransformer func(string) string

	// fieldsTransformer takes a list of field names and transforms them to the appropriate expected field names.
	fieldsTransformer func([]string) string
)

var (
	identityFn  = func(s string) string { return s }                            //nolint: gochecknoglobals
	fieldJoiner = func(flds []string) string { return strings.Join(flds, ",") } //nolint: gochecknoglobals
)

var deskObjectsWithFieldQuerySupport = datautils.NewSet( //nolint: gochecknoglobals
	"accounts", "tickets", "contacts")

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	switch c.moduleID { // nolint: exhaustive
	case providers.ModuleZohoDesk:
		// Desk uses lowercase field names with comma separation
		return c.buildModuleURL(config, deskAPIVersion, identityFn, fieldJoiner)
	default:
		// CRM uses capitalized field names with custom formatting
		return c.buildModuleURL(config, crmAPIVersion, naming.CapitalizeFirstLetter, func(flds []string) string {
			return strings.Join(flds, ",")
		})
	}
}

func (c *Connector) buildModuleURL(params common.ReadParams, apiVersion string,
	objTransformer objectNameTransformer, fldTransformer fieldsTransformer,
) (*urlbuilder.URL, error) {
	// Check if we're reading the next-page.
	if len(params.NextPage) > 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	objectName := c.transformedObjectName(params, objTransformer)

	url, err := c.getAPIURL(apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	c.constructIncrementalParams(url, params)

	fields := c.prepareFields(params, fldTransformer)

	if c.moduleID == providers.ModuleZohoCRM || (c.moduleID == providers.ModuleZohoDesk &&
		deskObjectsWithFieldQuerySupport.Has(params.ObjectName)) {
		url.WithQueryParam("fields", fields)
	}

	return url, nil
}

func (c *Connector) transformedObjectName(params common.ReadParams, transformer objectNameTransformer) string {
	if params.ObjectName != users && params.ObjectName != org {
		return transformer(params.ObjectName)
	}

	return params.ObjectName
}

func (c *Connector) prepareFields(params common.ReadParams, fieldTransformer fieldsTransformer) string {
	fieldSet := datautils.NewStringSet(params.Fields.List()...)

	if c.moduleID == providers.ModuleZohoDesk {
		c.ensureTimestampFields(fieldSet, params.ObjectName)
	}

	return fieldTransformer(fieldSet.List())
}

func (c *Connector) ensureTimestampFields(fieldSet datautils.StringSet, objectName string) {
	if objectsSortableByCreatedTime.Has(objectName) && !fieldSet.Has(createdTimeKey) {
		fieldSet.Add([]string{createdTimeKey})
	}

	if objectsSortablebyModifiedTime.Has(objectName) && !fieldSet.Has(modifiedTimeKey) {
		fieldSet.Add([]string{modifiedTimeKey})
	}
}

func (c *Connector) constructIncrementalParams(url *urlbuilder.URL, params common.ReadParams) {
	if c.moduleID != providers.ModuleZohoDesk {
		return
	}

	c.applySinceParam(url, params)
	url.WithQueryParam("limit", deskLimit)
}

func (c *Connector) applySinceParam(url *urlbuilder.URL, params common.ReadParams) {
	if params.Since.IsZero() {
		return
	}

	switch {
	case endpointsWithModifiedAfterParam.Has(params.ObjectName):
		url.WithQueryParam("modifiedAfter", params.Since.Format(time.RFC3339))
	case objectsSortableByCreatedTime.Has(params.ObjectName):
		url.WithQueryParam("sortBy", "-createdTime")
	case objectsSortablebyModifiedTime.Has(params.ObjectName):
		url.WithQueryParam("sortBy", "-modifiedTime")
	}
}
