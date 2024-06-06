package mockutils

import (
	"fmt"
	"net/http"
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
