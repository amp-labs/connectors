package providers

import (
	"github.com/amp-labs/connectors/common"
)

// HTTPOptionSpec declares an extra header or query parameter that a provider
// requires on every request beyond the primary auth credential, with the
// value sourced from a connection metadata field (see ProviderMetadata.Input).
// Specs are registered next to the provider's SetInfo call and resolved
// against connection metadata with ResolveHTTPOptions.
type HTTPOptionSpec struct {
	// In is where the option is attached on the request (header or query).
	In common.HTTPOptionIn

	// Key is the header or query parameter name, e.g. "ClientId".
	Key string

	// MetadataField is the connection metadata key holding the value,
	// matching the Name of a ProviderMetadata.Input item, e.g. "clientId".
	MetadataField string
}

//nolint:gochecknoglobals
var httpOptionSpecs = make(map[Provider][]HTTPOptionSpec)

// SetHTTPOptionSpecs registers the extra request options a provider requires.
// Like SetInfo, it is meant to be called from the provider's init.
func SetHTTPOptionSpecs(provider Provider, specs ...HTTPOptionSpec) {
	httpOptionSpecs[provider] = specs
}

// GetHTTPOptionSpecs returns the specs registered for the provider, if any.
func GetHTTPOptionSpecs(provider Provider) []HTTPOptionSpec {
	return httpOptionSpecs[provider]
}

// ResolveHTTPOptions resolves the provider's registered specs against
// connection metadata, producing options ready for NewClientParams.HTTPOptions
// or common.NewHTTPOptionsClient. Specs whose metadata field is absent or
// empty are skipped: connectors enforce their own metadata requirements, and
// failing client construction here would take down flows that never needed
// the option.
func ResolveHTTPOptions(provider Provider, metadata map[string]string) []common.HTTPOption {
	specs := httpOptionSpecs[provider]
	if len(specs) == 0 {
		return nil
	}

	opts := make([]common.HTTPOption, 0, len(specs))

	for _, spec := range specs {
		value := metadata[spec.MetadataField]
		if value == "" {
			continue
		}

		opts = append(opts, common.HTTPOption{
			In:    spec.In,
			Key:   spec.Key,
			Value: value,
		})
	}

	return opts
}
