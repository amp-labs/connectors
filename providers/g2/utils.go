package g2

import (
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	PathBuyerIntent        string = "buyer_intent"
	PathCompetitors        string = "competitors"
	PathDiscussions        string = "discussions"
	PathDownloads          string = "downloads"
	PathReviews            string = "reviews"
	PathProductScreenshots string = "product/screenshots"
	PathSnippets           string = "snippets"
	PathIntegrationReviews string = "integration_reviews"
	PathFeatures           string = "features"
	PathVideos             string = "videos"
	PathVideoReviews       string = "video_reviews"

	PathCategories         string = "categories"
	PathProducts           string = "products"
	PathCategoriesFeatures string = "categories/features"
	PathProductsFeatures   string = "products/features"
	PathProductMappings    string = "product_mappings"
	PathScreenshots        string = "screenshots"
	PathQuestions          string = "questions"
	PathVendors            string = "vendors"
)

// PathsConfig returns the appropriate path string based on the object name.
// For product-specific paths, it requires a productID and returns a path in the
// format "products/{productID}/{objectName}".
func PathsConfig(productID, objectName string) (string, error) {
	switch objectName {
	// Product-specific paths - require productID
	case PathBuyerIntent, PathCompetitors, PathDiscussions, PathReviews,
		PathProductScreenshots, PathSnippets, PathIntegrationReviews,
		PathFeatures, PathVideos, PathVideoReviews:
		return fmt.Sprintf("products/%s/%s", productID, objectName), nil

	// Paths - don't require productID
	case PathCategories, PathProducts, PathCategoriesFeatures,
		PathProductsFeatures, PathProductMappings, PathScreenshots,
		PathQuestions, PathVendors:
		return objectName, nil

	default:
		return "", common.ErrObjectNotSupported
	}
}

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	var (
		url *urlbuilder.URL
		err error
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

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, path)
	if err != nil {
		return nil, err
	}

	// Add fields query values
	if cfg.fieldsQuery != "" {
		url.WithQueryParam(cfg.fieldsQuery, strings.Join(params.Fields.List(), ","))
	}

	// Add page size query values
	if cfg.pageSizeQuery != "" {
		var pageSize = "100"
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

type ObjectConfig struct {
	fieldsQuery      string
	sinceQuery       string
	untilQuery       string
	pageSizeQuery    string
	maximumPerPage   string
	sinceValueFormat string
}

var readObjCfg = map[string]ObjectConfig{
	PathBuyerIntent: {
		fieldsQuery:      "dimensions",
		sinceQuery:       "dimension_filters[day_gteq]",
		pageSizeQuery:    "page[size]",
		maximumPerPage:   "100",
		sinceValueFormat: time.DateOnly,
	},
	PathCategories: {
		fieldsQuery:      "fields[categories]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathCompetitors: {
		fieldsQuery:    "fields[products]",
		pageSizeQuery:  "per",
		maximumPerPage: "50",
	},
	PathDiscussions: {
		fieldsQuery: "fields[discussions]",
	},
	PathDownloads: {
		fieldsQuery:      "fields[downloads]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathIntegrationReviews: { //needs incremental read live test
		fieldsQuery: "fields[integration_reviews]",
	},
	PathCategoriesFeatures: {
		fieldsQuery:      "fields[product_features]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathFeatures: {
		fieldsQuery:      "fields[product_features]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathProductsFeatures: {
		fieldsQuery:      "fields[product_features]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathProductMappings: {
		fieldsQuery:      "fields[product_mappings]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathVideos: {
		fieldsQuery:      "fields[product_videos]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathProducts: {
		fieldsQuery:   "fields[products]",
		pageSizeQuery: "page[size]",
	},
	PathQuestions: {
		fieldsQuery:      "fields[questions]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathReviews: {
		fieldsQuery:      "fields[reviews]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathScreenshots: {
		fieldsQuery:      "fields[screenshots]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathProductScreenshots: {
		fieldsQuery:      "fields[screenshots]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathSnippets: {
		fieldsQuery: "fields[snippets]",
	},
	PathVendors: {
		fieldsQuery:      "fields[vendors]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
	PathVideoReviews: {
		fieldsQuery:      "fields[video_reviews]",
		sinceQuery:       "filter[updated_at_gt]",
		untilQuery:       "filter[updated_at_lt]",
		sinceValueFormat: time.RFC3339,
	},
}

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node, "data").ArrayOptional(objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		links, err := jsonquery.New(node).ObjectOptional("links")
		if err != nil {
			return "", err
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
