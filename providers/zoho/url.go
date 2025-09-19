package zoho

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type (
	// objectNameTransformer takes an object and transfoms it to a standard zoho provider api name.
	objectNameTransformer func(string) string

	// objectNameTransformer takes a list of field names and transforms them to the appropriate expected field names.
	fieldsTrasformer func([]string) string
)

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	switch c.moduleID { // nolint: exhaustive
	case providers.ZohoDeskV2:
		return c.buildModuleURL(config, deskAPIVersion, func(s string) string { return s },
			func(flds []string) string { return strings.Join(flds, ",") })
	default:
		return c.buildModuleURL(config, crmAPIVersion, naming.CapitalizeFirstLetter, constructFieldNames)
	}
}

func (c *Connector) buildModuleURL(params common.ReadParams, apiVersion string,
	objTransfomer objectNameTransformer, fldTransformer fieldsTrasformer,
) (*urlbuilder.URL, error) {
	obj := params.ObjectName

	// Check if we're reading the next-page.
	if len(params.NextPage) > 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	// objects like users, org, org/currencies, __features,
	// uses lowecased object-names.
	if params.ObjectName != users && params.ObjectName != org {
		// Object names in ZohoCRM API are case sensitive.
		// Capitalizing the first character of object names to form correct URL.
		obj = objTransfomer(params.ObjectName)
	}

	url, err := c.getAPIURL(apiVersion, obj)
	if err != nil {
		return nil, err
	}

	if c.moduleID == providers.ZohoDeskV2 && params.ObjectName != "organizations" {
		url.WithQueryParam("orgId", c.orgID)
	}

	fields := fldTransformer(params.Fields.List())
	url.WithQueryParam("fields", fields)

	return url, nil
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
