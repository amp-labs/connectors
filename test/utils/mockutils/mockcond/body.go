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

type FieldIgnoreRule struct {
	name  string
	zooms []string
}

// IgnoreBodyField creates a rule that ignores the specified field path during body comparison.
//
// Examples:
//   - IgnoreBodyField("id")
//   - IgnoreBodyField("id", "profile", "user") -- ignores 'profile.user.id'.
func IgnoreBodyField(fieldName string, zoom ...string) FieldIgnoreRule {
	return FieldIgnoreRule{
		name:  fieldName,
		zooms: zoom,
	}
}

// BodyBytes returns a check that expects the request body to match the provided template bytes.
//
// If ignore rules are provided, they are applied when comparing JSON bodies.
func BodyBytes(expected []byte, rules ...FieldIgnoreRule) Check {
	return Body(string(expected), rules...)
}

// Body returns a check that expects the request body to match the provided template text.
//
// The body is compared as plain text, JSON, or XML. When JSON comparison is
// used, any provided ignore rules are applied to nested fields before matching.
func Body(expected string, rules ...FieldIgnoreRule) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		reader := r.Body

		body, err := io.ReadAll(reader)
		if err != nil {
			return false
		}

		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		textEquals := textBodyMatch(body, expected)
		jsonEquals := jsonBodyMatch(body, expected, rules)
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
func PermuteJSONBody(template string, slots PermuteSlots, rules ...FieldIgnoreRule) Condition {
	var build func(int, string) Condition

	build = func(i int, current string) Condition {
		// base case: all slots filled
		if i == len(slots) {
			return Body(current, rules...)
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

func jsonBodyMatch(actual []byte, expected string, rules []FieldIgnoreRule) bool {
	first := make(map[string]any)
	if err := json.Unmarshal(actual, &first); err != nil {
		return false
	}

	second := make(map[string]any)
	if err := json.Unmarshal([]byte(expected), &second); err != nil {
		return false
	}

	firstObj := applyIgnoreRules(first, rules)
	secondObj := applyIgnoreRules(second, rules)

	return reflect.DeepEqual(firstObj, secondObj)
}

func applyIgnoreRules(entity any, rules []FieldIgnoreRule) any {
	for _, rule := range rules {
		entity = removePath(entity, rulePath(rule))
	}

	return entity
}

func rulePath(rule FieldIgnoreRule) []string {
	path := make([]string, 0, 1+len(rule.zooms))
	path = append(path, rule.zooms...)
	path = append(path, rule.name)

	return path
}

func removePath(v any, path []string) any {
	if len(path) == 0 {
		return v
	}

	switch current := v.(type) {
	case map[string]any:
		key := path[0]
		if len(path) == 1 {
			delete(current, key)
			return current
		}

		next, ok := current[key]
		if !ok {
			return current
		}

		current[key] = removePath(next, path[1:])
		return current

	case []any:
		for i := range current {
			current[i] = removePath(current[i], path)
		}
		return current

	default:
		return v
	}
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

type PermuteSlots []PermuteSlot

type PermuteSlot struct {
	Name     string
	Values   []string
	NoQuotes bool // optional: auto JSON quote
}
