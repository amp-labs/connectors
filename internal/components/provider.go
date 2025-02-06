package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/providers"
)

// ProviderContext is a component that adds provider information to a connector.
type ProviderContext struct {
	provider     providers.Provider
	providerInfo *providers.ProviderInfo
	module       common.ModuleID
}

func NewProviderContext(
	p providers.Provider,
	module common.ModuleID,
	workspace string,
	metadata map[string]string,
) (*ProviderContext, error) {
	component := &ProviderContext{provider: p}
	metadata[catalogreplacer.VariableWorkspace] = workspace

	// TODO: Use module to get provider info
	providerInfo, err := providers.ReadInfo(p, paramsbuilder.NewCatalogVariables(metadata)...)
	if err != nil {
		return nil, err
	}

	component.providerInfo = providerInfo
	component.module = module

	return component, nil
}

func (p *ProviderContext) String() string {
	return p.provider + ".Connector"
}

func (p *ProviderContext) Provider() providers.Provider {
	return p.provider
}

func (p *ProviderContext) ProviderInfo() *providers.ProviderInfo {
	return p.providerInfo
}

func (p *ProviderContext) Module() common.ModuleID {
	return p.module
}
