package api3

import (
	"strings"

	"github.com/amp-labs/connectors/common/handy"
)

type PathMatcherStrategy interface {
	GivePathMatcher() PathMatcher
}

// AllowPathStrategy is a path matching strategy that will accept only those paths that matched the list.
// Others will be denied.
// You can use star symbol to create a wild matcher.
// Ex:
// Basic:	/v1/orders	- matches exact path
// Suffix:	*/batch		- matches paths ending with batch
// Prefix:	/v2/*		- matches paths starting with v2.
type AllowPathStrategy struct {
	Paths []string
}

// DenyPathStrategy is a path matching strategy that will deny only those paths that matched the list.
// Others will be allowed.
// You can use star symbol to create a wild matcher.
// Ex:
// Basic:	/v1/orders	- deny exact path
// Suffix:	*/batch		- deny paths ending with batch
// Prefix:	/v2/*		- deny paths starting with v2.
type DenyPathStrategy struct {
	Paths []string
}

type PathMatcher interface {
	IsPathMatching(path string) bool
}

func (s AllowPathStrategy) GivePathMatcher() PathMatcher { //nolint:ireturn
	return newStarRulePathResolver(s.Paths, func(matched bool) bool {
		return matched
	})
}

func (s DenyPathStrategy) GivePathMatcher() PathMatcher { //nolint:ireturn
	return newStarRulePathResolver(s.Paths, func(matched bool) bool {
		// if matched, deny instead.
		return !matched
	})
}

// This path resolver will report if path matches endpoint rule.
// Match can occur in 3 different ways,
// * exact value is inside the registry
// * or using star rule for
//   - prefix matching,
//   - suffix matching.
type starRulePathResolver struct {
	endpoints            handy.StringSet
	prefixes             []string
	suffixes             []string
	pathMatchingCallback func(hasMatched bool) bool
}

func newStarRulePathResolver(
	endpoints []string,
	pathMatchingCallback func(matched bool) bool,
) *starRulePathResolver {
	result := &starRulePathResolver{
		endpoints:            handy.NewStringSet(),
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

func (s starRulePathResolver) IsPathMatching(path string) bool {
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
