package components

import (
	"fmt"

	"github.com/amp-labs/connectors/providers"
)

// ProviderComponent is a component that adds provider information to a connector.
type ProviderComponent struct {
	provider     providers.Provider
	providerInfo *providers.ProviderInfo
}

func newProviderComponent(p providers.Provider, metadata map[string]string) (*ProviderComponent, error) {
	component := &ProviderComponent{provider: p}

	providerInfo, err := providers.ReadInfoMap(component.provider, metadata)
	if err != nil {
		return nil, err
	}

	component.providerInfo = providerInfo

	return component, nil
}

func (p *ProviderComponent) String() string {
	return fmt.Sprintf("%s.Connector", p.Provider)
}

func (p *ProviderComponent) Provider() providers.Provider {
	return p.provider
}

func (p *ProviderComponent) ProviderInfo() *providers.ProviderInfo {
	return p.providerInfo
}
