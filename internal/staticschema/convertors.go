package staticschema

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

var ErrObjectNotFound = errors.New("object not found")

// Select will look for object names under the module and will return metadata result for those objects.
// NOTE: empty module id is treated as core module.
func (r *Metadata) Select(
	moduleID common.ModuleID, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	moduleID = moduleIdentifier(moduleID)

	// Convert and return only listed objects
	module, ok := r.Modules[moduleID]
	if !ok {
		return nil, fmt.Errorf("%w: connector is using unknown module [%v]", common.ErrMissingModule, moduleID)
	}

	list := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: nil,
	}

	// Lookup each object under the module.
	for _, objectName := range objectNames {
		if v, ok := module.Objects[objectName]; ok {
			// move metadata from scrapper object to common object
			list.Result[objectName] = common.ObjectMetadata{
				DisplayName: v.DisplayName,
				FieldsMap:   v.FieldsMap,
			}
		} else {
			return nil, fmt.Errorf("%w: unknown object [%v]", ErrObjectNotFound, objectName)
		}
	}

	return list, nil
}
