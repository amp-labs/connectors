package reader

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/connectwise/internal/contactmodel"
	"github.com/amp-labs/connectors/providers/connectwise/internal/core"
	"github.com/amp-labs/connectors/providers/connectwise/internal/customfields"
	"github.com/spyzhov/ajson"
)

const DefaultPageSize = "1000"

func NextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}

// RecordTransformer returns a RecordTransformer that
//
//	(1) lifts customFields from the nested ConnectWise response into the top-level record map.
//	(2) lifts communication items of a contact as virtual fields.
func RecordTransformer(objectName string) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		if err = customfields.EnhanceReadRecord(node, root); err != nil {
			return nil, err
		}

		if objectName == core.ObjectNameContacts {
			if err = contactmodel.EnhanceReadRecord(node, root); err != nil {
				return nil, err
			}
		}

		return root, nil
	}
}
