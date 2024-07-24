package scrapper

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

var ErrObjectNotFound = errors.New("object not found")

func (r *ObjectMetadataResult) Select(objectNames []string) (*common.ListObjectMetadataResult, error) {
	list := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: nil,
	}

	// Convert and return only listed objects
	for _, objectName := range objectNames {
		name := naming.NewLowerString(objectName)
		if v, ok := r.Result[name]; ok {
			// move metadata from scrapper object to common object
			list.Result[name.String()] = common.ObjectMetadata{
				DisplayName: v.DisplayName,
				FieldsMap:   v.FieldsMap,
			}
		} else {
			return nil, fmt.Errorf("%w: unknown object [%v]", ErrObjectNotFound, name)
		}
	}

	return list, nil
}
