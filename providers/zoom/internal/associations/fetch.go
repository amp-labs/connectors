package associations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

// It retrieves a single resource and treats the entire JSON response as one record.
func fetchItem(ctx context.Context, base *core.Base, url *urlbuilder.URL, identifierField string) (recordItems, error) {
	if err := core.Wait(ctx); err != nil {
		return nil, err
	}

	resp, err := base.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		if isNotFound(err) {
			// No associations to return.
			return make(recordItems), nil
		}

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

// fetchPagedParticipants retrieves all participants from a paginated endpoint.
func fetchPagedParticipants(ctx context.Context, // nolint:cyclop,funlen
	base *core.Base,
	url *urlbuilder.URL,
) (recordItems, error) {
	records := make(recordItems)

	responseFieldName := "participants"
	identifierField := "user_id"
	// The counter is maintained across pages to ensure unique fallback IDs.
	counter := 0

	for {
		if err := core.Wait(ctx); err != nil {
			return nil, err
		}

		resp, err := base.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			if isNotFound(err) {
				// No associations to return.
				return make(recordItems), nil
			}

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

			// Use a synthetic ID when user_id is empty.
			// Zoom external participants often have many empty fields, but user_id
			// is generally available. The fallback prevents records with empty IDs
			// from overwriting each other in the map.
			if identifier == "" {
				identifier = fmt.Sprintf("external_user_%d", counter)
				counter++
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

func isNotFound(err error) bool {
	if errors.Is(err, common.ErrNotFound) {
		text := err.Error()
		if strings.Contains(text, "error code (3322)") ||
			strings.Contains(text, "error code (3001)") {
			// This meeting transcript (code 3322) or summary (code 3001) or participants (code 3001) does not exist.
			return true
		}
	}

	return false
}
