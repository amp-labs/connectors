package crm

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

const (
	readAPIVersion   = "api/v2"
	data             = "data"
	dealsObjectName  = "deals"
	productsFieldKey = "products"

	// ref: https://pipedrive.readme.io/docs/core-api-concepts-rate-limiting
	maxConcurrentProductFetch = 4 // No strong reason (Pipedrive has 10 requests per 2 seconds)
)

var supportsIncSync = datautils.NewSet("activities", "deals", "organizations", "persons") // nolint: gochecknoglobals

func (a *Adapter) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	result, err := common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(data),
		nextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
	if err != nil {
		return nil, err
	}

	if params.ObjectName == dealsObjectName && params.Fields.Has(productsFieldKey) {
		if err := a.enrichDealsWithProducts(ctx, result.Data); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// enrichDealsWithProducts fetches products for each deal concurrently and injects
// them into Fields["products"] on each row.
func (a *Adapter) enrichDealsWithProducts(ctx context.Context, rows []common.ReadResultRow) error {
	jobs := make([]simultaneously.Job, len(rows))

	for i := range rows {
		idx := i
		dealID := rows[i].Id

		jobs[idx] = func(ctx context.Context) error {
			products, err := a.fetchDealProducts(ctx, dealID)
			if err != nil {
				return fmt.Errorf("fetching products for deal %s: %w", dealID, err)
			}

			// we manually add these cause we have already parsed the response
			// at this stage.
			rows[idx].Fields[productsFieldKey] = products
			rows[idx].Raw[productsFieldKey] = products

			return nil
		}
	}

	return simultaneously.DoCtx(ctx, maxConcurrentProductFetch, jobs...)
}

// fetchDealProducts calls /api/v2/deals/{id}/products and returns the data array.
func (a *Adapter) fetchDealProducts(ctx context.Context, dealID string) ([]map[string]any, error) {
	url, err := urlbuilder.New(a.moduleInfo.BaseURL, readAPIVersion, dealsObjectName, dealID, productsFieldKey)
	if err != nil {
		return nil, err
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	result, err := common.UnmarshalJSON[records](resp)
	if err != nil {
		return nil, err
	}

	if result.Data == nil {
		return []map[string]any{}, nil
	}

	return result.Data, nil
}

func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	url, err := a.getAPIURL(readAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if supportsIncSync.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("updated_since", params.Since.Format(time.RFC3339))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("updated_until", params.Since.Format(time.RFC3339))
		}
	}

	return url, nil
}
