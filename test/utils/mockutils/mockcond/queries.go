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
	for param, values := range subset {
		superValues := make(map[string]bool)

		strList, ok := superset[param]
		if !ok {
			return false
		}

		for _, v := range strList {
			superValues[v] = true
		}

		for _, value := range values {
			if _, found := superValues[value]; !found {
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
