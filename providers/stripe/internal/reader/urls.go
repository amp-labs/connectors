package reader

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/stripe/internal/metadata"
)

// buildFirstPageReadURL constructs the API request URL for a Read operation.
// It applies:
//   - Pagination: adds "limit" query parameter
//   - Incremental reading: adds "created[gte]" for objects supporting incremental reads
//   - Object expansion: adds "expand[]" for fields of format:
//     => "$['line_items']['currency']"
//     => "$['line_items']['description']"
//     => "$['line_items']['...']"
//     => "$['source']['payment_intent']['customer']['id']"
//     => "$['source']['payment_intent']['id']"
//
// See [Stripe expand documentation](https://docs.stripe.com/expand#how-it-works).
func (s *Strategy) buildFirstPageReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := s.base.GetURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, strconv.Itoa(DefaultPageSize))
	url.WithQueryParam("limit", pageSize)

	if !params.Since.IsZero() && incrementalObjects.Has(params.ObjectName) {
		url.WithQueryParam("created[gte]", strconv.FormatInt(params.Since.Unix(), 10))
	}

	expandTargets := make(datautils.Set[string])

	for field := range params.Fields {
		// If the requested field supports expansion, add it to the expand list
		// so the provider can return the nested object in the response.
		queryParam := metadata.MakeExpandableQueryParam(params.ObjectName, field)
		if queryParam != "" {
			expandTargets.AddOne(queryParam)
		}
	}

	if len(expandTargets) != 0 {
		// Expand nested objects by adding "data.<field>" to expand[] query parameter
		url.WithQueryParamList("expand[]", expandTargets.List())
	}

	return url, nil
}
