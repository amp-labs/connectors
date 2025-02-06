package components

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/gobwas/glob"
)

// EndpointRegistry manages operation support for provider endpoints.
// It uses glob pattern matching to determine support levels for endpoints.
// Read more on how to write these patterns:
// https://github.com/gobwas/glob
type EndpointRegistry struct {
	patterns EndpointRegistryInput
}

// EndpointSupport defines support configuration for modules and their endpoints. Each key in this map
// is a module that supports an array of endpoints. Each endpoint is defined by a string and a support level.
// For example, you may define a root module that supports reading /users/* and /accounts/*, but only supports
// writing to /users/:id.
type EndpointRegistryInput map[common.ModuleID][]struct {
	Endpoint string
	Support  providers.Support
	glob     glob.Glob // Compiled pattern for matching
}

func NewEndpointRegistry(es EndpointRegistryInput) (*EndpointRegistry, error) {
	if err := precompileEndpoints(es); err != nil {
		return nil, fmt.Errorf("failed to compile endpoint patterns: %w", err)
	}

	return &EndpointRegistry{patterns: es}, nil
}

// Quick access to common support levels.
// nolint:gochecknoglobals
var (
	// DeleteSupport
	// TODO: For now, use bulkwrite.Delete as a stand-in for delete support.
	DeleteSupport = providers.Support{BulkWrite: providers.BulkWriteSupport{Delete: true}}
	ReadSupport   = providers.Support{Read: true}
	WriteSupport  = providers.Support{Write: true}

	NoSupport  = providers.Support{}
	AllSupport = providers.Support{
		BulkWrite: providers.BulkWriteSupport{
			Delete: true,
			Insert: true,
			Update: true,
			Upsert: true,
		},
		Proxy:     true,
		Read:      true,
		Subscribe: true,
		Write:     true,
	}
)

// GetSupport determines support for an endpoint by matching against registered patterns.
// Multiple patterns can match, in which case their support levels are combined.
func (p *EndpointRegistry) GetSupport(module common.ModuleID, path string) (*providers.Support, error) {
	endpoints, exists := p.patterns[module]
	if !exists {
		return &NoSupport, nil
	}

	support := providers.Support{}
	matched := false

	for _, endpoint := range endpoints {
		if endpoint.glob.Match(path) {
			matched = true

			mergeSupport(&support, endpoint.Support)
		}
	}

	if !matched {
		return &NoSupport, nil
	}

	return &support, nil
}

// mergeSupport combines two support configurations using OR operations.
func mergeSupport(base *providers.Support, additional providers.Support) {
	base.Read = base.Read || additional.Read
	base.Write = base.Write || additional.Write
	base.Proxy = base.Proxy || additional.Proxy
	base.Subscribe = base.Subscribe || additional.Subscribe
	base.BulkWrite.Delete = base.BulkWrite.Delete || additional.BulkWrite.Delete
	base.BulkWrite.Insert = base.BulkWrite.Insert || additional.BulkWrite.Insert
	base.BulkWrite.Update = base.BulkWrite.Update || additional.BulkWrite.Update
	base.BulkWrite.Upsert = base.BulkWrite.Upsert || additional.BulkWrite.Upsert
}

// precompileEndpoints compiles globs for all the URL endpoints in the support registry. The glob package recommends
// doing this for better performance.
func precompileEndpoints(registry EndpointRegistryInput) error {
	for _, endpoints := range registry {
		for i := range endpoints {
			if endpoints[i].glob != nil {
				continue
			}

			g, err := glob.Compile(endpoints[i].Endpoint)
			if err != nil {
				return err
			}

			endpoints[i].glob = g
		}
	}

	return nil
}
