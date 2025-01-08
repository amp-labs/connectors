package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

// ProviderComponent is a component that adds provider information to a connector.
type ProviderComponent struct {
	provider     providers.Provider
	providerInfo *providers.ProviderInfo
	module       common.ModuleID
}

func newProviderComponent(
	p providers.Provider,
	module common.ModuleID,
	metadata map[string]string,
) (*ProviderComponent, error) {
	component := &ProviderComponent{provider: p}

	// TODO: Use module to get provider info
	providerInfo, err := providers.ReadInfo(component.provider, paramsbuilder.NewCatalogVariables(metadata)...)
	if err != nil {
		return nil, err
	}

	component.providerInfo = providerInfo
	component.module = module

	return component, nil
}

func (p *ProviderComponent) String() string {
	return p.provider + ".Connector"
}

func (p *ProviderComponent) Provider() providers.Provider {
	return p.provider
}

func (p *ProviderComponent) ProviderInfo() *providers.ProviderInfo {
	return p.providerInfo
}

func (p *ProviderComponent) Module() common.ModuleID {
	return p.module
}
