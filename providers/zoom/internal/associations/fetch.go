package associations

import (
	"context"

	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

// It retrieves a single resource and treats the entire JSON response as one record.
func fetchItem(ctx context.Context, base *core.Base, url *urlbuilder.URL, identifierField string) (recordItems, error) {
	core.Wait()

	resp, err := base.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return make(recordItems), nil
	}

	record, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	if len(record) == 0 {
		return make(recordItems), nil
	}

	identifier, err := jsonquery.New(body).TextWithDefault(identifierField, "")
	if err != nil {
		return nil, err
	}

	return recordItems{identifier: record}, nil
}

// It retrieves a collection of resources from a paginated endpoint.
func fetchItems(ctx context.Context, // nolint:cyclop
	base *core.Base,
	url *urlbuilder.URL,
	responseFieldName, identifierField string,
) (recordItems, error) {
	records := make(recordItems)

	for {
		core.Wait()

		resp, err := base.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		body, ok := resp.Body()
		if !ok {
			return records, nil
		}

		array, err := jsonquery.New(body).ArrayOptional(responseFieldName)
		if err != nil {
			return nil, err
		}

		if len(array) == 0 {
			return records, nil
		}

		for _, item := range array {
			record, err := jsonquery.Convertor.ObjectToMap(item)
			if err != nil {
				return nil, err
			}

			identifier, err := jsonquery.New(item).TextWithDefault(identifierField, "")
			if err != nil {
				return nil, err
			}

			records[identifier] = record
		}

		nextPageToken, err := jsonquery.New(body).StrWithDefault("next_page_token", "")
		if err != nil {
			return nil, err
		}

		if nextPageToken == "" {
			return records, nil
		}

		url.WithQueryParam("next_page_token", nextPageToken)
	}
}
