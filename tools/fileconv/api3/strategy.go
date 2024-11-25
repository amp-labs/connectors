package api3

import (
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

// NewAllowPathStrategy produces a path matching strategy that will accept only those paths that matched the list.
// Others will be denied.
// You can use star symbol to create a wild matcher.
// Ex:
// Basic:	/v1/orders	- matches exact path
// Suffix:	*/batch		- matches paths ending with batch
// Prefix:	/v2/*		- matches paths starting with v2.
func NewAllowPathStrategy(paths []string) *StarRulePathResolver {
	return newStarRulePathResolver(paths, func(matched bool) bool {
		return matched
	})
}

// NewDenyPathStrategy produces a path matching strategy that will deny only those paths that matched the list.
// Others will be allowed.
// You can use star symbol to create a wild matcher.
// Ex:
// Basic:	/v1/orders	- deny exact path
// Suffix:	*/batch		- deny paths ending with batch
// Prefix:	/v2/*		- deny paths starting with v2.
func NewDenyPathStrategy(paths []string) *StarRulePathResolver {
	return newStarRulePathResolver(paths, func(matched bool) bool {
		// if matched, deny instead.
		return !matched
	})
}

type PathMatcher interface {
	IsPathMatching(path string) bool
}

// StarRulePathResolver will report if path matches endpoint rule.
// Match can occur in 3 different ways,
// * exact value is inside the registry
// * or using star rule for
//   - prefix matching,
//   - suffix matching.
type StarRulePathResolver struct {
	endpoints            datautils.StringSet
	prefixes             []string
	suffixes             []string
	contains             []string
	pathMatchingCallback func(hasMatched bool) bool
}

func newStarRulePathResolver(
	endpoints []string,
	pathMatchingCallback func(matched bool) bool,
) *StarRulePathResolver {
	result := &StarRulePathResolver{
		endpoints:            datautils.NewStringSet(),
		prefixes:             make([]string, 0),
		suffixes:             make([]string, 0),
		contains:             make([]string, 0),
		pathMatchingCallback: pathMatchingCallback,
	}

	for _, endpoint := range endpoints {
		if strings.HasPrefix(endpoint, "*") && strings.HasSuffix(endpoint, "*") {
			result.contains = append(result.contains, strings.Trim(endpoint, "*"))
		} else if rule, ok := strings.CutPrefix(endpoint, "*"); ok {
			result.suffixes = append(result.suffixes, rule)
		} else if rule, ok = strings.CutSuffix(endpoint, "*"); ok {
			result.prefixes = append(result.prefixes, rule)
		} else {
			result.endpoints.AddOne(endpoint)
		}
	}

	return result
}

func (s StarRulePathResolver) IsPathMatching(path string) bool {
	if s.endpoints.Has(path) {
		return s.pathMatchingCallback(true)
	}

	for _, middle := range s.contains {
		if stringsInTheMiddle(path, middle) {
			return s.pathMatchingCallback(true)
		}
	}

	for _, prefix := range s.prefixes {
		if strings.HasPrefix(path, prefix) {
			return s.pathMatchingCallback(true)
		}
	}

	for _, suffix := range s.suffixes {
		if strings.HasSuffix(path, suffix) {
			return s.pathMatchingCallback(true)
		}
	}

	return s.pathMatchingCallback(false)
}

func stringsInTheMiddle(text, substr string) bool {
	parts := strings.Split(text, substr)
	if len(parts) != 2 {
		return false
	}

	// There must be text preceding and succeeding substring to be in the middle.
	return len(parts[0]) != 0 && len(parts[1]) != 0
}
