package capsule

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/capsule/metadata"
)

const apiVersion = "v2"

// https://developer.capsulecrm.com/v2/operations/Custom_Field#listFields
func (c *Connector) getCustomFieldsURLFor(objectName string) (*urlbuilder.URL, error) {
	objectName = mapObjectAlias(objectName)

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, "fields/definitions")
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.Module(), objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
}

func (c *Connector) getWriteURL(objectName string, id string) (*urlbuilder.URL, error) {
	objectName = mapObjectAlias(objectName)

	if len(id) == 0 {
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, id)
}

func (c *Connector) getDeleteURL(objectName string, id string) (*urlbuilder.URL, error) {
	objectName = mapObjectAlias(objectName)

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, id)
}

// In future kases will be renamed to projects.
// The API still doesn't recognize "projects" as REST resource name.
// https://developer.capsulecrm.com/v2/models/project
func mapObjectAlias(objectName string) string {
	if objectName == objectNameProjects {
		objectName = "kases"
	}

	return objectName
}
