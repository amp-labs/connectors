package providers

import (
	"github.com/amp-labs/connectors/internal/generated"
)

// ================================================================================
// Contains critical provider configuration (using types from types.gen.go)
// Populated by the init() functions in the various .go provider files.
// Please don't add providers here, add them in the appropriate provider file.
// ================================================================================

// pre-allocate the catalog to the number of providers we think we have.
// This count is automatically updated by the catalog/json.go script,
// which runs anytime the catalog is updated.
//
// NOTE: There's no need to ever manually update the count. Even if it's
// off by a few (which isn't a big deal), it'll get corrected next time
// the script runs. So don't worry about it.
var catalog = make(CatalogType, generated.ProviderCount) // nolint:gochecknoglobals

func getCatalog() *CatalogWrapper {
	return &CatalogWrapper{
		Catalog:   catalog,
		Timestamp: generated.Timestamp,
	}
}
