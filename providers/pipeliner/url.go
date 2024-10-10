package pipeliner

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"github.com/amp-labs/connectors/internal/deep/dpvars"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var _ deep.ObjectURLResolver = customURLBuilder{}

func newURLBuilder(
	data *dpvars.ConnectorData[parameters, *dpvars.EmptyMetadataVariables],
) *customURLBuilder {
	return &customURLBuilder{
		data: data,
	}
}

type customURLBuilder struct {
	data *dpvars.ConnectorData[parameters, *dpvars.EmptyMetadataVariables]
}

func (f customURLBuilder) FindURL(method deep.Method, baseURL, objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(baseURL,
		"api/v100/rest/spaces/", f.data.Workspace, "/entities", objectName)
}

func (f customURLBuilder) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "objectUrlResolver",
		Constructor: newURLBuilder,
		Interface:   new(deep.ObjectURLResolver),
	}
}
