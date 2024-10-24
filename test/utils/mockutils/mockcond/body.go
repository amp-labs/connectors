package mockcond

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// BodyBytes returns a check expecting body to match template bytes.
func BodyBytes(expected []byte) Check {
	return Body(string(expected))
}

// Body returns a check expecting body to match template text.
func Body(expected string) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		reader := r.Body

		body, err := io.ReadAll(reader)
		if err != nil {
			return false
		}

		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		textEquals := textBodyMatch(body, expected)
		jsonEquals := jsonBodyMatch(body, expected)

		return textEquals || jsonEquals
	}
}

func jsonBodyMatch(actual []byte, expected string) bool {
	first := make(map[string]any)
	if err := json.Unmarshal(actual, &first); err != nil {
		return false
	}

	second := make(map[string]any)
	if err := json.Unmarshal([]byte(expected), &second); err != nil {
		return false
	}

	return reflect.DeepEqual(first, second)
}

func textBodyMatch(actual []byte, expected string) bool {
	first := stringCleaner(string(actual), []string{"\n", "\t"})
	second := stringCleaner(expected, []string{"\n", "\t"})

	return first == second
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
