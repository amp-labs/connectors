package mockcond

import (
	"net/http"
	"net/url"
	"time"
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

// QueryParamTimeApprox returns a Check that validates a query parameter represents
// a time approximately equal to a reference time. It uses a fixed tolerance of
// 5 minutes. This is a convenience wrapper around QueryParamTimeDelta.
//
// key: the name of the query parameter to check
// refTime: the reference time to compare against
// timeFormat: the expected time format in the query parameter (e.g., "2006-01-02")
//
// Example:
//
//   fixedNow := time.Date(2026, 1, 29, 12, 0, 0, 0, time.UTC)
//   QueryParamTimeApprox("to", fixedNow, "2006-01-02")
//
func QueryParamTimeApprox(key string, refTime time.Time, timeFormat string) Check {
	return QueryParamTimeDelta(key, refTime, timeFormat, 5*time.Minute)
}

// QueryParamTimeDelta returns a Check that validates a query parameter represents
// a time within a specified delta (tolerance) of a reference time.
//
// key: the name of the query parameter to check
// refTime: the reference time to compare against
// timeFormat: the expected time format in the query parameter (e.g., "2006-01-02", time.RFC3339)
// delta: the maximum allowed absolute difference between the query param time and refTime
//
// The check succeeds if:
//   1) the query parameter exists
//   2) it can be parsed using the provided timeFormat
//   3) the absolute difference between the parsed time and refTime is <= delta
//
// Example:
//
//   fixedNow := time.Date(2026, 1, 29, 12, 0, 0, 0, time.UTC)
//   QueryParamTimeDelta("to", fixedNow, "2006-01-02", 10*time.Minute)
//
func QueryParamTimeDelta(key string, refTime time.Time, timeFormat string, delta time.Duration) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		q := r.URL.Query().Get(key)
		if q == "" {
			return false
		}

		// Parse the query parameter using the expected format
		t, err := time.Parse(timeFormat, q)
		if err != nil {
			return false
		}

		// Compute absolute difference
		diff := t.Sub(refTime)
		if diff < 0 {
			diff = -diff
		}

		return diff <= delta
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
