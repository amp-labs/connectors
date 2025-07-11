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

// PathMatcher categorizes paths based on specific criteria.
type PathMatcher interface {
	IsPathMatching(path string) bool
}

// DefaultPathMatcher matches any URL path without restrictions.
type DefaultPathMatcher struct{}

func (DefaultPathMatcher) IsPathMatching(path string) bool {
	return true
}

// AndPathMatcher combines multiple path matchers.
// A path matches only if all matchers in the list agree on the match.
type AndPathMatcher []PathMatcher

func (m AndPathMatcher) IsPathMatching(path string) bool {
	// Find the first not matching. Conclude doesn't match the combined goal.
	for _, matcher := range m {
		if matcher == nil {
			continue
		}

		if !matcher.IsPathMatching(path) {
			return false
		}
	}

	return true
}

// OrPathMatcher combines multiple path matchers.
// A path matches if at least one matcher in the list agrees on the match.
type OrPathMatcher []PathMatcher

func (m OrPathMatcher) IsPathMatching(path string) bool {
	// Find the first not matching. Conclude doesn't match the combined goal.
	for _, matcher := range m {
		if matcher == nil {
			continue
		}

		if matcher.IsPathMatching(path) {
			return true
		}
	}

	return false
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
		pathMatchingCallback: pathMatchingCallback,
	}

	for _, endpoint := range endpoints {
		if rule, ok := strings.CutPrefix(endpoint, "*"); ok {
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

type IDPathIgnorer struct{}

func (IDPathIgnorer) IsPathMatching(path string) bool {
	// There should be no slashes, curly brackets.
	// If there are we are dealing with nester resources. Those are to be ignored.
	return !strings.Contains(path, "{")
}

type NestedIDPathIgnorer struct{}

func (NestedIDPathIgnorer) IsPathMatching(path string) bool {
	// Remove the last URL part.
	parts := strings.Split(path, "/")
	parts = parts[:len(parts)-1]
	path = strings.Join(parts, "/")

	// We get the large prefix which may have variate parts. It shouldn't.
	return !strings.Contains(path, "{")
}
