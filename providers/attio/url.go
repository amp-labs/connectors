package attio

import (
	"strings"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

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
