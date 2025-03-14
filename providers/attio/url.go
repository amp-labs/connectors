package attio

import (
	"strings"
)

// Relative URL for retrieving metadata for standard and custom objects in Attio.
func (c *Connector) getObjectsURL(objName string) string {
	relativeURL := strings.Join([]string{"objects", objName, "attributes"}, "/")

	return relativeURL
}

// Relative URL for retrieving display Name for standard and custom objects in Attio.
func (c *Connector) getObjects(objName string) string {
	relativeURL := strings.Join([]string{"objects", objName}, "/")

	return relativeURL
}
