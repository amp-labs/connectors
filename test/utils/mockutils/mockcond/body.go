package mockcond

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common/xquery"
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
		xmlEquals := xmlBodyMatch(body, expected)

		return textEquals || jsonEquals || xmlEquals
	}
}

// PermuteJSONBody returns a Condition expecting the request body to match
// a template where array ordering may vary.
//
// The template may contain one or more placeholders using the syntax
// "%name". Each placeholder corresponds to a PermuteSlot and is replaced
// with every permutation of the slot's Values. All permutations across
// all slots are generated automatically, allowing tests to ignore
// ordering differences in serialized arrays.
//
// Example:
//
//	`{"properties":[%properties],"filters":[%filters]}`
//
// Multiple placeholders may be used in the same template. Each occurrence
// of a placeholder receives the same permutation ordering.
func PermuteJSONBody(template string, slots ...PermuteSlot) Condition {
	var build func(int, string) Condition

	build = func(i int, current string) Condition {
		// base case: all slots filled
		if i == len(slots) {
			return Body(current)
		}

		slot := slots[i]
		placeholder := "%" + slot.Name

		return Permute(
			func(order []string) Condition {
				withQuotes := !slot.NoQuotes
				replacement := render(order, withQuotes)

				next := strings.ReplaceAll(
					current,
					placeholder,
					replacement,
				)

				// recurse into next permutation layer
				return build(i+1, next)
			},
			slot.Values...,
		)
	}

	return build(0, template)
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

func xmlBodyMatch(actual []byte, expected string) bool {
	first, err := xquery.NewXML(actual)
	if err != nil {
		return false
	}

	second, err := xquery.NewXML([]byte(expected))
	if err != nil {
		return false
	}

	return first.EqualsIgnoreOrder(second)
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

func render(values []string, quote bool) string {
	if quote {
		out := make([]string, len(values))
		for i, v := range values {
			out[i] = strconv.Quote(v)
		}
		return strings.Join(out, ",")
	}

	return strings.Join(values, ",")
}

type PermuteSlot struct {
	Name     string
	Values   []string
	NoQuotes bool // optional: auto JSON quote
}
