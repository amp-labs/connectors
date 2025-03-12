package capsule

import (
	"github.com/amp-labs/connectors/internal/components/endpoints"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
)

var writeIdentifiers = endpoints.ResponseIdentifierRegistry{ // nolint:gochecknoglobals
	staticschema.RootModuleID: datautils.NewDefaultMap(map[string]endpoints.JSONPath{
		// TODO
	}, func(objectName string) endpoints.JSONPath {
		return endpoints.NewJSONPath("")
	}),
}

var createEndpoints = endpoints.OperationRegistry{ // nolint:gochecknoglobals
	staticschema.RootModuleID: datautils.NewDefaultMap(map[string]endpoints.OperationSpec{
		// TODO
	}, func(objectName string) endpoints.OperationSpec {
		return endpoints.OperationSpec{
			Method: "",
			Path:   "",
		}
	}),
}

var updateEndpoints = endpoints.OperationRegistry{ // nolint:gochecknoglobals
	staticschema.RootModuleID: datautils.NewDefaultMap(map[string]endpoints.OperationSpec{
		// TODO
	}, func(objectName string) endpoints.OperationSpec {
		return endpoints.OperationSpec{
			Method: "",
			Path:   "",
		}
	}),
}

var deleteEndpoints = endpoints.OperationRegistry{ // nolint:gochecknoglobals
	staticschema.RootModuleID: datautils.NewDefaultMap(map[string]endpoints.OperationSpec{
		// TODO
	}, func(objectName string) endpoints.OperationSpec {
		return endpoints.OperationSpec{
			Method: "",
			Path:   "",
		}
	}),
}
