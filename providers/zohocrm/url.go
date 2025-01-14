package zohocrm

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// Check if we're reading the next-page.
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	// Object names in ZohoCRM API are case sensitive.
	// Capitalizing the first character of object names to form correct URL.
	obj := naming.CapitalizeFirstLetterEveryWord(config.ObjectName)

	url, err := c.getAPIURL(obj)
	if err != nil {
		return nil, err
	}

	fields := constructFieldNames(config.Fields.List())
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
