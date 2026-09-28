package reducers

import (
	"testing"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestShortestNameFromURL(t *testing.T) {
	tests := []struct {
		name         string
		namingFromTo map[string]string
	}{
		{
			name: "Single URL",
			namingFromTo: map[string]string{
				"/garden/baskets/apples": "apples",
			},
		},
		{
			name: "Simple collision",
			namingFromTo: map[string]string{
				"/garden/apples":  "garden/apples",
				"/baskets/apples": "baskets/apples",
			},
		},
		{
			name: "Deep collision",
			namingFromTo: map[string]string{
				"/inventory/locations/shelves/items": "inventory/locations/shelves/items",
				"/baskets/locations/shelves/items":   "baskets/locations/shelves/items",
			},
		},
		{
			name: "Partial collision",
			namingFromTo: map[string]string{
				"/garden/baskets/apples/seeds":   "baskets/apples/seeds",
				"/garden/inventory/apples/seeds": "inventory/apples/seeds",
				"/store/aisle/peaches":           "peaches",
			},
		},
		{
			name: "Collision with different lengths",
			namingFromTo: map[string]string{
				"/garden/baskets/apples": "baskets/apples",
				"/inventory/apples":      "inventory/apples",
			},
		},
		{
			name: "Trailing slashes",
			namingFromTo: map[string]string{
				"/garden/baskets/": "garden/baskets",
				"/storage/baskets": "storage/baskets",
			},
		},
		{
			name: "Basic progressive disambiguation",
			namingFromTo: map[string]string{
				"/orders/line-items/products":      "line-items/products",
				"/orders/line-items/discounts":     "discounts",
				"/orders/shipments/tracking":       "shipments/tracking",
				"/warehouse/members":               "warehouse/members",
				"/warehouse/items/tracking":        "items/tracking",
				"/warehouse/items/products":        "items/products",
				"/teams/members":                   "teams/members",
				"/teams/channels/messages":         "messages",
				"/teams/channels/files":            "files",
				"/post-office/deliveries/tracking": "deliveries/tracking",
				"/post-office/deliveries/packages": "packages",
				"/post-office/products":            "post-office/products",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := datautils.Map[string, string](tt.namingFromTo).Keys()
			input := makeSchemas(urls)
			output := ShortestNameFromUrl(input)
			actual := getURLToObjectName(output)

			result := testutils.NewCompareResult()
			result.Assert("Mapping from URL to ObjectName", tt.namingFromTo, actual)
			result.Validate(t, "ShortestNameFromURL")
		})
	}
}

func TestShortestNameFromURLWithFullFallback(t *testing.T) {
	tests := []struct {
		name         string
		namingFromTo map[string]string
	}{
		{
			name: "Single URL",
			namingFromTo: map[string]string{
				"/garden/baskets/apples": "apples",
			},
		},
		{
			name: "Simple collision (full fallback)",
			namingFromTo: map[string]string{
				"/garden/apples":  "garden/apples",
				"/baskets/apples": "baskets/apples",
			},
		},
		{
			name: "Deep collision (full fallback)",
			namingFromTo: map[string]string{
				"/inventory/locations/shelves/items": "inventory/locations/shelves/items",
				"/baskets/locations/shelves/items":   "baskets/locations/shelves/items",
			},
		},
		{
			name: "Mixed unique and colliding",
			namingFromTo: map[string]string{
				"/garden/baskets/apples":    "garden/baskets/apples",
				"/inventory/baskets/apples": "inventory/baskets/apples",
				"/store/aisle/peaches":      "peaches",
			},
		},
		{
			name: "Trailing slashes",
			namingFromTo: map[string]string{
				"/garden/baskets/": "garden/baskets",
				"/storage/baskets": "storage/baskets",
			},
		},
		{
			name: "Collision with different lengths (full fallback)",
			namingFromTo: map[string]string{
				"/garden/baskets/apples": "garden/baskets/apples",
				"/inventory/apples":      "inventory/apples",
			},
		},
		{
			name: "Full fallback disambiguation",
			namingFromTo: map[string]string{
				"/orders/line-items/discounts":     "discounts",
				"/orders/line-items/products":      "orders/line-items/products",
				"/orders/shipments/tracking":       "orders/shipments/tracking",
				"/post-office/deliveries/packages": "packages",
				"/post-office/deliveries/tracking": "post-office/deliveries/tracking",
				"/post-office/products":            "post-office/products",
				"/teams/channels/files":            "files",
				"/teams/channels/messages":         "messages",
				"/teams/members":                   "teams/members",
				"/warehouse/items/products":        "warehouse/items/products",
				"/warehouse/items/tracking":        "warehouse/items/tracking",
				"/warehouse/members":               "warehouse/members",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := datautils.Map[string, string](tt.namingFromTo).Keys()
			input := makeSchemas(urls)
			output := ShortestNameFromUrlWithFullFallback(input)
			actual := getURLToObjectName(output)

			result := testutils.NewCompareResult()
			result.Assert("Mapping from URL to ObjectName", tt.namingFromTo, actual)
			result.Validate(t, "ShortestNameFromURLWithFullFallback")
		})
	}
}

// makeSchemas wraps URLs into spec.Schema objects.
func makeSchemas(urls []string) []spec.Schema {
	schemas := make([]spec.Schema, len(urls))
	for idx, url := range urls {
		schemas[idx] = spec.Schema{
			UrlPath: url,
		}
	}
	return schemas
}

// getURLToObjectName extracts mapping from UrlPath to ObjectName from spec.Schema objects.
func getURLToObjectName(schemas []spec.Schema) map[string]string {
	registry := make(map[string]string)
	for _, schema := range schemas {
		registry[schema.UrlPath] = schema.ObjectName
	}

	return registry
}
