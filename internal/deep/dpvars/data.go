package dpvars

import (
	"errors"

	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// ErrInvalidMetadataVariableType it is returned when generics used by connector implementor didn't match of what
// is circulating among connector component dependencies. deep.Connector should receive matching MetadataVariables.
var ErrInvalidMetadataVariableType = errors.New("injected metadata variables do not match connector type")

// ConnectorData is a concrete representation of connector parameters and paramsbuilder.Metadata.
// You can specify this connector component as an argument to the connector constructor.
// By default, metadata is dpvars.EmptyMetadataVariables, which is an empty holder.
// If deep connector has defined a concrete metadata type which actually holds some data, you should use that type.
// Ex:
//
// (default) - connector or custom component by default should inject ConnectorData with empty metadata type.
//
//			type Connector struct {
//				Data dpvars.ConnectorData[parameters, *dpvars.EmptyMetadataVariables]
//				deep.Clients
//	 		...
//			}
//
// (inject with custom metadata) - your custom component now holds reference to ConnectorData.
//
//	func someConnectorComponentConstructor(
//		data *dpvars.ConnectorData[parameters, *CustomMetadataVariables], ... ) *myConnectorComponent {
//		return &myConnectorComponent{data:data}
//	}
type ConnectorData[P paramsbuilder.ParamAssurance, D MetadataVariables] struct {
	Workspace string
	Module    string
	Metadata  D
}

func newConnectorDescriptor[P paramsbuilder.ParamAssurance, D MetadataVariables](
	parameters *Parameters[P],
	metadataVariables MetadataVariables,
) (*ConnectorData[P, D], error) {
	data := new(ConnectorData[P, D])

	if holder, ok := parameters.Params.(paramsbuilder.WorkspaceHolder); ok {
		workspace := holder.GiveWorkspace()
		data.Workspace = workspace.Name
	}

	if holder, ok := parameters.Params.(paramsbuilder.ModuleHolder); ok {
		module := holder.GiveModule()
		data.Module = module.Selection.Path()
	}

	if holder, ok := parameters.Params.(paramsbuilder.MetadataHolder); ok {
		metadata := holder.GiveMetadata()
		metadataVariables.FromMap(metadata.Map)

		data.Metadata, ok = metadataVariables.(D)
		if !ok {
			return nil, ErrInvalidMetadataVariableType
		}
	}

	return data, nil
}

func (c ConnectorData[P, D]) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ConnectorData,
		Constructor: newConnectorDescriptor[P, D],
	}
}
