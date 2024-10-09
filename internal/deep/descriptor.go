package deep

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type ConnectorDescriptor[P paramsbuilder.ParamAssurance, D MetadataVariables] struct {
	Workspace string
	Module    string
	Metadata  D
}

func newConnectorDescriptor[P paramsbuilder.ParamAssurance, D MetadataVariables](
	parameters *Parameters[P],
	metadataVariables MetadataVariables,
) *ConnectorDescriptor[P, D] {

	descr := new(ConnectorDescriptor[P, D])

	if holder, ok := parameters.Params.(paramsbuilder.WorkspaceHolder); ok {
		workspace := holder.GiveWorkspace()
		descr.Workspace = workspace.Name
	}

	if holder, ok := parameters.Params.(paramsbuilder.ModuleHolder); ok {
		module := holder.GiveModule()
		descr.Module = module.Name
	}

	if holder, ok := parameters.Params.(paramsbuilder.Metadata); ok {
		metadata := holder.GiveMetadata()
		metadataVariables.FromMap(metadata.Map)
		descr.Metadata, ok = metadataVariables.(D)
		if !ok {
			// TODO return an error, connector descriptor should have the same type as metadata variables.
		}
	}

	return descr
}

func (c ConnectorDescriptor[P, D]) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "connectorDescriptor",
		Constructor: newConnectorDescriptor[P, D],
	}
}
