package capsule

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/capsule/metadata"
)

const apiVersion = "v2"

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.Module(), objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
}

func (c *Connector) getWriteURL(objectName string, id string) (*urlbuilder.URL, error) {
	if objectName == objectNameProjects {
		objectName = "kases"
	}

	if len(id) == 0 {
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, id)
}

func (c *Connector) getDeleteURL(objectName string, id string) (*urlbuilder.URL, error) {
	if objectName == objectNameProjects {
		objectName = "kases"
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, id)
}
