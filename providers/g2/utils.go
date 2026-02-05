package g2

import (
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	pathBuyerIntent        string = "buyer_intent"
	pathBuyerIntentSandBox string = "sandbox/buyer_intent"
	pathCompetitors        string = "competitors"
	pathDiscussions        string = "discussions"
	pathDownloads          string = "downloads"
	pathReviews            string = "reviews"
	pathProductScreenshots string = "product/screenshots"
	pathSnippets           string = "snippets"
	pathIntegrationReviews string = "integration_reviews"
	pathFeatures           string = "features"
	pathVideos             string = "videos"
	pathVideoReviews       string = "video_reviews"

	pathCategories         string = "categories"
	pathProducts           string = "products"
	pathCategoriesFeatures string = "categories/features"
	pathProductsFeatures   string = "products/features"
	pathProductMappings    string = "product_mappings"
	pathScreenshots        string = "screenshots"
	pathQuestions          string = "questions"
	pathVendors            string = "vendors"

	maxPageSize = "100"
)

type ObjectConfig struct {
	fieldsQuery      string
	sinceQuery       string
	untilQuery       string
	pageSizeQuery    string
	maximumPerPage   string
	sinceValueFormat string
}

var readObjCfg = map[string]ObjectConfig{ // nolint: gochecknoglobals
	pathBuyerIntent: {
		fieldsQuery:      "dimensions",
		sinceQuery:       "dimension_filters[time_gteq]",
		pageSizeQuery:    "page[size]",
		maximumPerPage:   "100",
		sinceValueFormat: time.RFC3339,
	},
	pathBuyerIntentSandBox: {
		fieldsQuery:      "dimensions",
		sinceQuery:       "dimension_filters[time_gteq]",
		pageSizeQuery:    "page[size]",
		maximumPerPage:   "50",
		sinceValueFormat: time.RFC3339,
	},
	pathCategories: {
		fieldsQuery:      "fields[categories]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathCompetitors: {
		fieldsQuery:    "fields[products]",
		pageSizeQuery:  "per",
		maximumPerPage: "50",
	},
	pathDiscussions: {
		fieldsQuery: "fields[discussions]",
	},
	pathDownloads: {
		fieldsQuery:      "fields[downloads]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathIntegrationReviews: { // needs incremental read live test
		fieldsQuery: "fields[integration_reviews]",
	},
	pathCategoriesFeatures: {
		fieldsQuery:      "fields[product_features]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathFeatures: {
		fieldsQuery:      "fields[product_features]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathProductsFeatures: {
		fieldsQuery:      "fields[product_features]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathProductMappings: {
		fieldsQuery:      "fields[product_mappings]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathVideos: {
		fieldsQuery:      "fields[product_videos]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathProducts: {
		fieldsQuery:   "fields[products]",
		pageSizeQuery: "page[size]",
	},
	pathQuestions: {
		fieldsQuery:      "fields[questions]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathReviews: {
		fieldsQuery:      "fields[reviews]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathScreenshots: {
		fieldsQuery:      "fields[screenshots]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathProductScreenshots: {
		fieldsQuery:      "fields[screenshots]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathSnippets: {
		fieldsQuery: "fields[snippets]",
	},
	pathVendors: {
		fieldsQuery:      "fields[vendors]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	pathVideoReviews: {
		fieldsQuery:      "fields[video_reviews]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
}

var buyerIntents = datautils.NewStringSet(pathBuyerIntent, pathBuyerIntentSandBox) // nolint: gochecknoglobals

// PathsConfig returns the appropriate path string based on the object name.
// For product-specific paths, it requires a productID and returns a path in the
// format "products/{productID}/{objectName}".
func PathsConfig(productID, objectName string) (string, error) {
	switch objectName {
	// Product-specific paths - require productID
	case pathBuyerIntent, pathCompetitors, pathDiscussions, pathReviews,
		pathProductScreenshots, pathSnippets, pathIntegrationReviews,
		pathFeatures, pathVideos, pathVideoReviews:
		return fmt.Sprintf("products/%s/%s", productID, objectName), nil

	// Paths - don't require productID
	case pathCategories, pathProducts, pathCategoriesFeatures,
		pathProductsFeatures, pathProductMappings, pathScreenshots,
		pathQuestions, pathVendors:
		return objectName, nil
	case pathBuyerIntentSandBox:
		return fmt.Sprintf("products/%s/buyer_intent", productID), nil
	default:
		return "", common.ErrObjectNotSupported
	}
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) { // nolint: cyclop,funlen
	var (
		url            *urlbuilder.URL
		err            error
		restAPIVersion = "api/v2"
	)

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return url, nil
	}

	path, err := PathsConfig(c.productId, params.ObjectName)
	if err != nil {
		return nil, err
	}

	cfg, exists := readObjCfg[params.ObjectName]
	if !exists {
		return nil, common.ErrObjectNotSupported
	}

	if params.ObjectName == pathBuyerIntentSandBox {
		restAPIVersion = "api/sandbox"
	}

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, path)
	if err != nil {
		return nil, err
	}

	// Add fields query values
	if cfg.fieldsQuery != "" {
		dimensions := make(datautils.StringSet)
		dimensions.Add(params.Fields.List())

		// If a user is reading buyer_intent we need the time field for incremental read sync.
		if !dimensions.Has("time") && (buyerIntents.Has(params.ObjectName)) { //nolint:lll
			dimensions.Add([]string{"time"})
		}

		// dimensions never takes id as a query param, thus we remove it from the list of fields we require from the API.
		// we will get this field from the raw records, as it's always returned.
		if dimensions.Has("id") {
			dimensions.Remove("id")
		}

		url.WithQueryParam(cfg.fieldsQuery, strings.Join(dimensions.List(), ","))
	}

	// Add page size query values
	if cfg.pageSizeQuery != "" {
		pageSize := maxPageSize

		if cfg.maximumPerPage != "" {
			pageSize = cfg.maximumPerPage
		}

		url.WithQueryParam(cfg.pageSizeQuery, pageSize)
	}

	if !params.Since.IsZero() && cfg.sinceQuery != "" {
		url.WithQueryParam(cfg.sinceQuery, params.Since.Format(cfg.sinceValueFormat))
	}

	return url, nil
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		links, err := jsonquery.New(node).ObjectOptional("links")
		if err != nil {
			return "", err
		}

		if links == nil {
			return "", nil
		}

		nextURL, err := jsonquery.New(links).StringOptional("next")
		if err != nil {
			return "", err
		}

		if nextURL == nil {
			return "", err
		}

		return *nextURL, nil
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
