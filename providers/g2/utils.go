package g2

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

const (
	PathBuyerIntent        = "buyer_intent"
	PathCompetitors        = "competitors"
	PathDiscussions        = "discussions"
	PathReviews            = "reviews"
	PathProductScreenshots = "product/screenshots"
	PathSnippets           = "snippets"
	PathIntegrationReviews = "integration_reviews"
	PathFeatures           = "features"
	PathVideos             = "videos"
	PathVideoReviews       = "video_reviews"

	PathCategories         = "categories"
	PathProducts           = "products"
	PathCategoriesFeatures = "categories/features"
	PathProductsFeatures   = "products/features"
	PathProductMappings    = "product_mappings"
	PathScreenshots        = "screenshots"
	PathQuestions          = "questions"
	PathVendors            = "vendors"
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
