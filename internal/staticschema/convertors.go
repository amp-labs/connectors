package staticschema

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// Select will look for object names under the module and will return metadata result for those objects.
// NOTE: empty module id is treated as root module.
func (m *Metadata[F]) Select(
	moduleID common.ModuleID, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	moduleID = moduleIdentifier(moduleID)

	// Convert and return only listed objects
	module, ok := m.Modules[moduleID]
	if !ok {
		return nil, fmt.Errorf("%w: connector is using unknown module [%v]", common.ErrMissingModule, moduleID)
	}

	list := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// Lookup each object under the module.
	for _, objectName := range objectNames {
		if v, ok := module.Objects[objectName]; ok {
			// move metadata from scrapper object to common object
			list.Result[objectName] = *v.getObjectMetadata()
		} else {
			return nil, fmt.Errorf("%w: unknown object [%v]", common.ErrObjectNotSupported, objectName)
		}
	}

	return list, nil
}

// SelectOne reads one object metadata from the static file.
// NOTE: empty module id is treated as root module.
func (m *Metadata[F]) SelectOne(
	moduleID common.ModuleID, objectName string,
) (*common.ObjectMetadata, error) {
	moduleID = moduleIdentifier(moduleID)

	// Convert and return only listed objects
	module, ok := m.Modules[moduleID]
	if !ok {
		return nil, fmt.Errorf("%w: connector is using unknown module [%v]", common.ErrMissingModule, moduleID)
	}

	if v, ok := module.Objects[objectName]; ok {
		return v.getObjectMetadata(), nil
	}

	return nil, fmt.Errorf("%w: unknown object [%v]", common.ErrObjectNotSupported, objectName)
}
