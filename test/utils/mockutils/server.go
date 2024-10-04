package mockutils

import (
	"fmt"
	"io"
	"net/http"
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

func WriteBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}

func bodiesMatch(reader io.ReadCloser, expected string) bool {
	body, err := io.ReadAll(reader)
	if err != nil {
		return false
	}

	return string(body) == stringCleaner(expected, []string{"\n", "\t"})
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
