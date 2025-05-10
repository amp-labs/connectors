package endpoints

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
)

// OperationRegistryFromStaticSchema creates OperationRegistry suitable for connector's READ operation.
func OperationRegistryFromStaticSchema[F staticschema.FieldMetadataMap, C any](
	metadata *staticschema.Metadata[F, C],
) OperationRegistry {
	registry := make(map[common.ModuleID]map[string]OperationSpec)

	for moduleID, module := range metadata.Modules {
		operations := make(map[string]OperationSpec)

		for objectName, object := range module.Objects {
			operations[objectName] = OperationSpec{
				Path: module.Path + object.URLPath,
			}
		}

		registry[moduleID] = operations
	}

	return NewOperationRegistry(http.MethodGet, registry, nil)
}
