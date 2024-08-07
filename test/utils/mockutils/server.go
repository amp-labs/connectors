package mockutils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func RespondNoContentForMethod(w http.ResponseWriter, r *http.Request, methodName string) {
	RespondToMethod(w, r, methodName, func() {
		w.WriteHeader(http.StatusNoContent)
	})
}

func RespondToMethod(w http.ResponseWriter, r *http.Request, methodName string, onSuccess func()) {
	// if method is not as expected we return error code so the test will fail
	// and with response payload which will be a helpful message for debugging
	if r.Method == methodName {
		// if method is matching execute callback
		onSuccess()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		WriteBody(w, fmt.Sprintf(`{
			"error": {
				"code": "from test",
				"message": "test server expected %v request"
			}}`, methodName))
	}
}

func RespondToHeader(w http.ResponseWriter, r *http.Request, header http.Header, onSuccess func()) {
	// if some headers are missing we return error code so the test will fail
	if missingHeader, ok := headerIsSubset(r.Header, header); ok {
		// if method is matching headers
		onSuccess()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		WriteBody(w, fmt.Sprintf(`{
			"error": {
				"code": "from test",
				"message": "test server mismatching [%v] header"
			}}`, missingHeader))
	}
}

func RespondToBody(w http.ResponseWriter, r *http.Request, body string, onSuccess func()) {
	if ok := bodiesMatch(r.Body, body); ok {
		// if method is matching bodies
		onSuccess()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		WriteBody(w, `{
			"error": {
				"code": "from test",
				"message": "test server mismatching bodies"
			}}`)
	}
}

func RespondToQueryParameters(w http.ResponseWriter, r *http.Request, queries url.Values, onSuccess func()) {
	// if some query parameters are mismatching return error code so the test will fail
	if queryParam, ok := queryParamsAreSubset(r.URL.Query(), queries); ok {
		// if method is matching headers
		onSuccess()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		WriteBody(w, fmt.Sprintf(`{
			"error": {
				"code": "from test",
				"message": "test server mismatching [%v] query parameter"
			}}`, queryParam))
	}
}

func RespondToMissingQueryParameters(w http.ResponseWriter, r *http.Request, missingQueries []string, onSuccess func()) {
	// if at least one query parameter exists return error code so the test will fail
	if queryParam, ok := queryParamsMissing(r.URL.Query(), missingQueries); ok {
		// if method is matching headers
		onSuccess()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		WriteBody(w, fmt.Sprintf(`{
			"error": {
				"code": "from test",
				"message": "test server found [%v] query parameter"
			}}`, queryParam))
	}
}

func WriteBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}

func headerIsSubset(superset, subset http.Header) (string, bool) {
	for name, values := range subset {
		superValues := make(map[string]bool)
		for _, v := range superset.Values(name) {
			superValues[v] = true
		}
		// every value of this header must be part of superset
		for _, value := range values {
			if _, found := superValues[value]; !found {
				return name, false
			}
		}
	}

	return "", true
}

func queryParamsAreSubset(superset, subset url.Values) (string, bool) {
	for param, values := range subset {
		superValues := make(map[string]bool)

		strings, ok := superset[param]
		if !ok {
			return param, false
		}

		for _, v := range strings {
			superValues[v] = true
		}

		for _, value := range values {
			if _, found := superValues[value]; !found {
				return param, false
			}
		}
	}

	return "", true
}

func bodiesMatch(reader io.ReadCloser, expected string) bool {
	body, err := io.ReadAll(reader)
	if err != nil {
		return false
	}

	return string(body) == stringCleaner(expected, []string{"\n", "\t"})
}

func queryParamsMissing(superset url.Values, missing []string) (string, bool) {
	for _, param := range missing {
		if superset.Has(param) {
			// query was found, while should be missing
			return param, false
		}
	}

	return "", true
}

func stringCleaner(text string, toRemove []string) string {
	rules := make(map[string]string)
	for _, remove := range toRemove {
		rules[remove] = ""
	}

	return stringReplacer(text, rules)
}

func stringReplacer(text string, rules map[string]string) string {
	for from, to := range rules {
		text = strings.ReplaceAll(text, from, to)
	}

	return text
}
