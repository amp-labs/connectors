package endpoints

import (
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// ResponseDataLocator uses API specifications to determine how to locate various data in an HTTP response.
type ResponseDataLocator struct {
	providerContext    *components.ProviderContext
	identifierRegistry ResponseIdentifierRegistry
}

func NewResponseDataLocator(
	providerContext *components.ProviderContext, identifierRegistry ResponseIdentifierRegistry,
) *ResponseDataLocator {
	return &ResponseDataLocator{
		providerContext:    providerContext,
		identifierRegistry: identifierRegistry,
	}
}

// ExtractRecordID retrieves the record ID from the response node,
// with the path varying based on the object name.
func (l ResponseDataLocator) ExtractRecordID(node *ajson.Node, objectName string) (string, error) {
	path := l.identifierRegistry[l.providerContext.Module()].Get(objectName)

	return jsonquery.New(node, path.nested...).TextWithDefault(path.key, "")
}
