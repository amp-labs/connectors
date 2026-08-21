package mappings

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// nolint:gochecknoglobals
var (
	objectsBotConnector  = make(objectsRegistry)
	objectsUserConnector = make(objectsRegistry)
)

type objectsRegistry map[string]Object

// Adds object definition or merges existing info properties.
func (r objectsRegistry) add(objectName string, object Object) {
	current, ok := r[objectName]
	if !ok {
		r[objectName] = object

		return
	}

	if object.readListInfo != nil {
		current.readListInfo = object.readListInfo
	}

	if object.readItemInfo != nil {
		current.readItemInfo = object.readItemInfo
	}

	if object.writeCreateInfo != nil {
		current.writeCreateInfo = object.writeCreateInfo
	}

	if object.writeUpdateInfo != nil {
		current.writeUpdateInfo = object.writeUpdateInfo
	}

	if object.deleteInfo != nil {
		current.deleteInfo = object.deleteInfo
	}

	r[objectName] = current
}

func getObject(provider providers.Provider, objectName string) (*Object, error) {
	switch provider {
	case providers.Slack:
		object, ok := objectsBotConnector[objectName]
		if !ok {
			return nil, common.ErrObjectNotSupported
		}

		return &object, nil
	default:
		return nil, common.ErrInvalidImplementation
	}
}
