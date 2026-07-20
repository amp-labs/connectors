package reader

import (
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// makeGetRecords creates a NodeRecordsFunc that extracts records from Stripe's API response.
// It retrieves the array field containing the list of records (e.g., "data" for most objects).
func makeGetRecords(responseFieldName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

// Pagination is implemented as follows:
//   - Check the response to determine if there are more items to retrieve.
//   - If additional items exist, extract the ID of the last item from the current page.
//   - Use this ID to query the next page, starting after the last item ID from the current page.
//
// For more details, refer to the documentation:
// https://docs.stripe.com/api/pagination?lang=curl
func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		hasMore, err := jsonquery.New(node).BoolWithDefault("has_more", false)
		if err != nil {
			return "", err
		}

		if !hasMore {
			return "", nil
		}

		data, err := jsonquery.New(node).ArrayOptional("data")
		if err != nil {
			return "", err
		}

		if len(data) == 0 {
			return "", nil
		}

		lastElement := data[len(data)-1]

		lastItemID, err := jsonquery.New(lastElement).StringOptional("id")
		if err != nil {
			return "", err
		}

		if lastItemID == nil {
			return "", nil
		}

		url.WithQueryParam("starting_after", *lastItemID)

		return url.String(), nil
	}
}

func fieldsSelector(node *ajson.Node, fields []string) (map[string]any, string, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, "", err
	}

	identifier, err := jsonquery.New(node).StringRequired("id")
	if err != nil {
		return nil, "", err
	}

	customFields, err := getCustomFields(node)
	if err != nil {
		return nil, "", err
	}

	selected := readhelper.SelectFields(root, datautils.NewSetFromList(fields))
	maps.Copy(selected, customFields)

	return selected, identifier, nil
}
