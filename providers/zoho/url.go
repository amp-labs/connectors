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
	timeLayout     = "2006-01-02T15:04:05.000Z"
	articlesObject = "articles"
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

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	switch c.moduleID { // nolint: exhaustive
	case providers.ZohoDesk:
		// Desk uses lowercase field names with comma separation
		return c.buildModuleURL(config, deskAPIVersion, identityFn, fieldJoiner)
	default:
		// CRM uses capitalized field names with custom formatting
		return c.buildModuleURL(config, crmAPIVersion, naming.CapitalizeFirstLetter, constructFieldNames)
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
	if params.ObjectName != articlesObject {
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

	if c.moduleID == providers.ZohoDesk {
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
	if c.moduleID != providers.ZohoDesk {
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

// Zoho field names typically start with a capital letter.
// For fields with multiple words, underscores are used to separate them.
// This function converts field names into a format that the API accepts.
func constructFieldNames(flds []string) string {
	cpdFlds := make([]string, len(flds))

	for idx, fld := range flds {
		// id is used and attached to the field parameter as is.
		if strings.ToLower(fld) == "id" {
			cpdFlds[idx] = fld

			continue
		}

		// Some fields end with `__s`, and the `s` should not be capitalized,
		// so we strip it first and then reattach it after capitalizing all the other words
		if strings.HasSuffix(fld, "__s") {
			fld = capitalizeSegments(fld[:len(fld)-3])
			fld += "__s"
			cpdFlds[idx] = fld
		} else {
			cpdFlds[idx] = capitalizeSegments(fld)
		}
	}

	return strings.Join(cpdFlds, ",")
}

func capitalizeSegments(fld string) string {
	// This maps fields to the unique available fields.
	mappedObject, ok := uniqueFields[strings.ToLower(fld)]
	if ok {
		return mappedObject
	}

	// Most Fields are in the structure XXX_XXXX (Full_Name).
	// thus we capitalize first letter of individual substrings.
	// Split the field by `_` and capitalize the individual segments.
	segments := strings.Split(fld, "_")
	for idx, seg := range segments {
		seg = naming.CapitalizeFirstLetterEveryWord(seg)
		// Update the segment to it's capitalized string.
		segments[idx] = seg
	}

	return strings.Join(segments, "_")
}
