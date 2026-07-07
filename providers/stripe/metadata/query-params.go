package metadata

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/readhelper"
)

/*
MakeExpandableQueryParam builds a Stripe `expand[]` query parameter for a field in JSON path format.
Regular fields can be passed and result in empty query parameter.

The connector only has partial knowledge of the nested object graph, so it can
only infer the deepest valid expansion path from the registry of known expandable
fields. When the requested path includes unsupported segments, the function trims
the query at the last resolvable expandable object.

Example:
(balance_transactions, $['source']['payment_intent']['customer']['id'])

	(balance_transactions, source) -> can expand
	(source, payment_intent)       -> unknown
	(payment_intent, customer)     -> can expand
	(customer, id)                 -> unknown

Therefore, the expand query must include everything up to customer:

	data.source.payment_intent.customer

Stripe expansion has a maximum nesting depth of 4 levels, so this helper stops
as soon as the path would exceed that limit. If the requested field path cannot
be matched to at least one expandable field, the function returns an empty string.

Docs:
https://docs.stripe.com/expand#multiple-levels

Example error returned by Stripe when the limit is exceeded:

	{
	    "error": {
	        "code": "property_expansion_max_depth",
	        "message": "You cannot expand more than 4 levels of a property.
						Property: data.source.payment_intent.customer.discount",
	        "request_log_url": "",
	        "type": "invalid_request_error"
	    }
	}
*/
func MakeExpandableQueryParam(objectName, field string) string {
	keys := readhelper.ParseJSONPath(field)
	if len(keys) == 0 {
		return ""
	}

	// We only have partial knowledge of the nested object graph, so we infer the
	// longest valid expansion path we can safely build from the registry.
	type pair struct {
		// resource is the Stripe object currently used to resolve the next field.
		// A resource may be known by more than one name, for example singular and plural.
		resource string

		// field is the next segment in the requested JSON path.
		// It may refer to either a primitive field or another nested Stripe object.
		field string

		// expandable reports whether the field can be expanded on the current resource.
		// This is resolved against the expandable-fields registry.
		expandable bool
	}

	// Partition the requested path into resource/field pairs.
	// The first segment is resolved relative to the root object.
	currentResource := objectName
	lastExpandablePair := 0
	pairs := make([]pair, len(keys))

	for index, key := range keys {
		expandable := expandableFields.Exists(currentResource, key)
		if !expandable && index == 0 {
			// If the root field is not expandable, the path cannot be turned into a
			// valid Stripe expand parameter.
			return ""
		}

		pairs[index] = pair{
			resource:   currentResource,
			field:      key,
			expandable: expandable,
		}

		currentResource = key

		if expandable {
			lastExpandablePair = index
		}
	}

	// If the final object itself is expandable, keep the path at that object.
	// This lets the caller expand the full object instead of stopping at a parent.
	if expandableFields.Has(currentResource) {
		lastExpandablePair = len(pairs) - 1
	}

	// Stripe list responses require the `data.` prefix in expand paths.
	// We construct the query from the root and stop before exceeding the 4-level limit.
	const maxPropertyLevel = 4

	var queryParamBuilder strings.Builder
	queryParamBuilder.WriteString("data")

	for i := 0; i <= lastExpandablePair; i++ {
		if i+1 == maxPropertyLevel {
			// Stripe enforces a maximum expansion depth of 4 levels.
			// If the requested path exceeds that limit, return an empty query parameter
			// value instead of generating a truncated query that would not yield the desired field.
			// The connector will skip this field.
			return ""
		}

		queryParamBuilder.WriteString(".")
		queryParamBuilder.WriteString(pairs[i].field)
	}

	return queryParamBuilder.String()
}

func (f ExpandableFieldsDef) Has(resourceName string) bool {
	possibleObjectNames := createPossibleObjectNames(resourceName)

	for _, objectName := range possibleObjectNames {
		if _, ok := f[objectName]; ok {
			return true
		}
	}

	return false
}

func (f ExpandableFieldsDef) Exists(resourceName, fieldName string) bool {
	possibleObjectNames := createPossibleObjectNames(resourceName)
	queryParam := fmt.Sprintf("data.%v", fieldName)

	for _, objectName := range possibleObjectNames {
		if f.queryParamExists(objectName, queryParam) {
			return true
		}
	}

	return false
}

func (f ExpandableFieldsDef) queryParamExists(objectName, queryParam string) bool {
	queryParams, ok := f[objectName]
	if !ok {
		return false
	}

	return queryParams.Has(queryParam)
}

// createPossibleObjectNames returns the resource name as provided and a pluralized
// variant.
//
// Stripe resource names and our internal naming do not always match exactly.
// In the common case, the API may refer to a singular object while the connector
// works with a plural form, so we try both.
//
// NOTE: If we later need special-case mappings for irregular names, this is the place to add them.
func createPossibleObjectNames(resourceName string) []string {
	return []string{
		// Try the resource name as provided.
		resourceName,
		// Some Stripe resources may be referenced in singular form, while we index by plural.
		naming.NewPluralString(resourceName).String(),
	}
}
