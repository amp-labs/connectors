package mockcond

import (
	"net/http"
	"net/url"
)

func QueryParam(key string, value ...string) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		return queryParamsAreSubset(r.URL.Query(), url.Values{
			key: value,
		})
	}
}

func QueryParamsMissing(keys ...string) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		return queryParamsMissing(r.URL.Query(), keys)
	}
}

func queryParamsAreSubset(superset, subset url.Values) bool {
	for param, expectedValues := range subset {
		actualValues := make(map[string]bool)

		values, ok := superset[param]
		if !ok {
			return false
		}

		for _, v := range values {
			actualValues[v] = true
		}

		if len(actualValues) != len(expectedValues) {
			// The query param lists must be of the same size.
			return false
		}

		for _, value := range expectedValues {
			if _, found := actualValues[value]; !found {
				return false
			}
		}
	}

	return true
}

func queryParamsMissing(superset url.Values, missing []string) bool {
	for _, param := range missing {
		if superset.Has(param) {
			// query was found, while should be missing
			return false
		}
	}

	return true
}
