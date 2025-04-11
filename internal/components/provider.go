package components

import (
	"fmt"

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
	provider providers.Provider,
	module common.ModuleID,
	workspace string,
	metadata map[string]string,
) (*ProviderContext, error) {
	pctx := &ProviderContext{provider: provider}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	metadata[catalogreplacer.VariableWorkspace] = workspace

	// TODO: Use module to get provider info
	providerInfo, err := providers.ReadInfo(provider, paramsbuilder.NewCatalogVariables(metadata)...)
	if err != nil {
		return nil, err
	}

	pctx.providerInfo = providerInfo
	pctx.module = module

	return pctx, nil
}

func (p *ProviderContext) String() string {
	return fmt.Sprintf("%v.Connector[%v]", p.provider, p.module)
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
