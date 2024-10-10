package dpvars

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type ConnectorData[P paramsbuilder.ParamAssurance, D MetadataVariables] struct {
	Workspace string
	Module    string
	Metadata  D
}

func newConnectorDescriptor[P paramsbuilder.ParamAssurance, D MetadataVariables](
	parameters *Parameters[P],
	metadataVariables MetadataVariables,
) *ConnectorData[P, D] {
	data := new(ConnectorData[P, D])

	if holder, ok := parameters.Params.(paramsbuilder.WorkspaceHolder); ok {
		workspace := holder.GiveWorkspace()
		data.Workspace = workspace.Name
	}

	if holder, ok := parameters.Params.(paramsbuilder.ModuleHolder); ok {
		module := holder.GiveModule()
		data.Module = module.Name
	}

	if holder, ok := parameters.Params.(paramsbuilder.MetadataHolder); ok {
		metadata := holder.GiveMetadata()
		metadataVariables.FromMap(metadata.Map)

		data.Metadata, ok = metadataVariables.(D)
		if !ok {
			// TODO return an error, connector descriptor should have the same type as metadata variables.
		}
	}

	return data
}

func (c ConnectorData[P, D]) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "connectorDescriptor",
		Constructor: newConnectorDescriptor[P, D],
	}
}
