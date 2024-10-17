package mockcond

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// Body returns a check expecting body to match template text.
func Body(expectedBody string) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		reader := r.Body

		body, err := io.ReadAll(reader)
		if err != nil {
			return false
		}

		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		a := stringCleaner(string(body), []string{"\n", "\t"})
		b := stringCleaner(expectedBody, []string{"\n", "\t"})
		match := a == b

		return match
	}
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
