package attio

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// A trailing slash is added to the relative URL to ensure proper concatenation of dynamic values.
//
// Relative URL for retrieving metadata for standard and custom objects in Attio.
func (c *Connector) getObjectAttributesURL(objName string) (*urlbuilder.URL, error) {
	return c.ModuleClient.URL("objects", objName, "attributes")
}

// Relative URL for retrieving display Name for standard and custom objects in Attio.
func (c *Connector) getObjectsURL(objName string) (*urlbuilder.URL, error) {
	return c.ModuleClient.URL("objects", objName)
}

// Relative URL for retrieving the options in the attributes.
func (c *Connector) getOptionsURL(objName, attributeID string) (*urlbuilder.URL, error) {
	return c.ModuleClient.URL("objects", objName, "attributes", attributeID, "options")
}

// Relative URL for retrieving standard and custom object read URL.
func (c *Connector) getObjectReadURL(objName string) (*urlbuilder.URL, error) {
	return c.ModuleClient.URL("objects", objName, "records", "query")
}

// Relative URL for retrieving standarad and custom object write URL.
func (c *Connector) getObjectWriteURL(objName string) (*urlbuilder.URL, error) {
	return c.ModuleClient.URL("objects", objName, "records")
}
