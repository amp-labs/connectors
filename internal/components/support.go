package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/gobwas/glob"
)

// ProviderEndpointSupport defines the support for endpoints of a provider. All matches are merged to provide the final
// support level for an endpoint/object. Explicit denies aren't considered *for now*, because this answers
// 'What do we want to allow?'.
type ProviderEndpointSupport struct {
	registry map[common.ModuleID][]EndpointSupport
}

type EndpointSupport struct {
	Endpoint string
	Support  providers.Support

	// Compiled glob from the endpoint string for pattern matching.
	glob glob.Glob
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

func NewProviderEndpointSupport(mp map[common.ModuleID][]EndpointSupport) (*ProviderEndpointSupport, error) {
	if err := precompileEndpoints(mp); err != nil {
		return nil, err
	}

	return &ProviderEndpointSupport{registry: mp}, nil
}

// nolint:cyclop
func (pr *ProviderEndpointSupport) GetSupport(module common.ModuleID, path string) (*providers.Support, error) {
	if endpoints, ok := pr.registry[module]; ok {
		support := providers.Support{}

		for _, esupport := range endpoints {
			if esupport.glob == nil {
				// We need to compile the glob.
				g, err := glob.Compile(esupport.Endpoint)
				if err != nil {
					return nil, err
				}

				esupport.glob = g
			}

			// There might be multiple endpoint matches, so add matched support to the final object.
			if esupport.glob.Match(path) {
				// There's better ways to do this, but for now, this works.
				support.Read = support.Read || esupport.Support.Read
				support.Write = support.Write || esupport.Support.Write
				support.Proxy = support.Proxy || esupport.Support.Proxy
				support.Subscribe = support.Subscribe || esupport.Support.Subscribe

				support.BulkWrite.Delete = support.BulkWrite.Delete || esupport.Support.BulkWrite.Delete
				support.BulkWrite.Insert = support.BulkWrite.Insert || esupport.Support.BulkWrite.Insert
				support.BulkWrite.Update = support.BulkWrite.Update || esupport.Support.BulkWrite.Update
				support.BulkWrite.Upsert = support.BulkWrite.Upsert || esupport.Support.BulkWrite.Upsert
			}
		}

		return &support, nil
	}

	// The module wasn't found in the registry. Reject - we don't support it.
	return &NoSupport, nil
}

// precompileEndpoints compiles globs for all the URL endpoints in the support registry. The glob package recommends
// doing this for better performance.
func precompileEndpoints(registry map[common.ModuleID][]EndpointSupport) error {
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
