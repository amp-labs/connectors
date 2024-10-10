package pipeliner

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var _ deep.URLResolver = customURLBuilder{}

func newURLBuilder(
	data *deep.ConnectorData[parameters, *deep.EmptyMetadataVariables],
	clients *deep.Clients,
) *customURLBuilder {
	return &customURLBuilder{
		data:    data,
		clients: clients,
	}
}

type customURLBuilder struct {
	data    *deep.ConnectorData[parameters, *deep.EmptyMetadataVariables]
	clients *deep.Clients
}

func (f customURLBuilder) FindURL(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(f.clients.BaseURL(),
		"api/v100/rest/spaces/", f.data.Workspace, "/entities", objectName)
}

func (f customURLBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "urlResolver",
		Constructor: newURLBuilder,
		Interface:   new(deep.URLResolver),
	}
}
