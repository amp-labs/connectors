package attio

import (
	"strings"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

// A trailing slash is added to the relative URL to ensure proper concatenation of dynamic values.
//
// Relative URL for retrieving metadata for standard and custom objects in Attio.
func (c *Connector) getObjectAttributesURL(objName string) (*urlbuilder.URL, error) {
	relativeURL := strings.Join([]string{"objects", objName, "attributes"}, "/")

	return urlbuilder.New(c.BaseURL, apiVersion, relativeURL)
}

// Relative URL for retrieving display Name for standard and custom objects in Attio.
func (c *Connector) getObjectsURL(objName string) (*urlbuilder.URL, error) {
	relativeURL := strings.Join([]string{"objects", objName}, "/")

	return urlbuilder.New(c.BaseURL, apiVersion, relativeURL)
}

// Relative URL for retrieving standard and custom object read URL.
func (c *Connector) getObjectReadURL(objName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, "objects", objName, "records", "query")
}

// Relative URL for retrieving standarad and custom object write URL.
func (c *Connector) getObjectWriteURL(objName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, "objects", objName, "records")
}
