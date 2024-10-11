package dpread

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/spyzhov/ajson"
)

const defaultNodeField = "data"

// ResponseLocator is a type of Responder which looks for object under some field name.
// By default, it expects list to be under "data" field. You can use callback to customize,
// have conditional branching based on Object.
// Some REST APIs have inconsistent response naming, which can be addressed here.
//
// Ex: ObjectName ==> FieldName
// * orders - orders
// * coupons - items
// * favourites - my_bookmarks
//
// Consider using handy.DefaultMap to handle exceptions.
type ResponseLocator struct {
	// Locate should return the fieldName where desired list of Objects is located. By default, it looks under "data".
	Locate func(config common.ReadParams, node *ajson.Node) string
	// Process is optional.
	// Use it if the desired fields are nested and you need to flatten them.
	// Also, any field reshuffling, pruning can happen here.
	Process func(arr []*ajson.Node) ([]map[string]any, error)
}

func (l ResponseLocator) GetRecordsFunc(config common.ReadParams) (common.RecordsFunc, error) {
	return func(node *ajson.Node) ([]map[string]any, error) {
		fieldName := defaultNodeField
		if l.Locate != nil {
			fieldName = l.Locate(config, node)
		}

		arr, err := jsonquery.New(node).Array(fieldName, false)
		if err != nil {
			return nil, err
		}

		if l.Process != nil {
			return l.Process(arr)
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}, nil
}

func (l ResponseLocator) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ReadObjectLocator,
		Constructor: handy.PtrReturner(l),
		Interface:   new(Responder),
	}
}
