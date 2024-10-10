package pipeliner

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var _ deep.URLResolver = customURLBuilder{}

func newURLBuilder(
	data *deep.ConnectorData[parameters, *deep.EmptyMetadataVariables],
) *customURLBuilder {
	return &customURLBuilder{
		data: data,
	}
}

type customURLBuilder struct {
	data *deep.ConnectorData[parameters, *deep.EmptyMetadataVariables]
}

func (f customURLBuilder) FindURL(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(baseURL,
		"api/v100/rest/spaces/", f.data.Workspace, "/entities", objectName)
}

func (f customURLBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "urlResolver",
		Constructor: newURLBuilder,
		Interface:   new(deep.URLResolver),
	}
}
