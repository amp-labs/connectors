package filters

import (
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

type Rule struct {
	raw     string
	pattern string
	used    bool
}

type Rules []*Rule

func NewRules(paths ...string) Rules {
	rules := make(Rules, len(paths))
	for index, value := range paths {
		rules[index] = &Rule{
			raw:  value,
			used: false,
		}
	}

	return rules
}

func (r Rules) UnusedRules() []string {
	rules := make([]string, 0)

	for _, rule := range r {
		if !rule.used {
			rules = append(rules, rule.raw)
		}
	}

	return rules
}

// NewAllowPathStrategy produces a path matching strategy that will accept only those paths that matched the list.
// Others will be denied.
// You can use star symbol to create a wild matcher.
// Ex:
// Basic:	/v1/orders	- matches exact path
// Suffix:	*/batch		- matches paths ending with batch
// Prefix:	/v2/*		- matches paths starting with v2.
func NewAllowPathStrategy(rules Rules) *StarRulePathResolver { // nolint:funcorder
	return newStarRulePathResolver(rules, func(matched bool) bool {
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
func NewDenyPathStrategy(rules Rules) *StarRulePathResolver { // nolint:funcorder
	return newStarRulePathResolver(rules, func(matched bool) bool {
		// if matched, deny instead.
		return !matched
	})
}

// PathMatcher categorizes paths based on specific criteria.
type PathMatcher interface {
	IsPathMatching(path string) bool
}

// KeepAll matches any URL path without restrictions.
type KeepAll struct{}

func (KeepAll) IsPathMatching(path string) bool {
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
	endpoints            datautils.Map[string, *Rule]
	prefixes             datautils.Map[string, *Rule]
	suffixes             datautils.Map[string, *Rule]
	pathMatchingCallback func(hasMatched bool) bool
}

func newStarRulePathResolver(
	rules Rules,
	pathMatchingCallback func(matched bool) bool,
) *StarRulePathResolver {
	result := &StarRulePathResolver{
		endpoints:            make(datautils.Map[string, *Rule]),
		prefixes:             make(datautils.Map[string, *Rule]),
		suffixes:             make(datautils.Map[string, *Rule]),
		pathMatchingCallback: pathMatchingCallback,
	}

	for _, rule := range rules {
		if pattern, ok := strings.CutPrefix(rule.raw, "*"); ok {
			result.suffixes[rule.raw] = rule
			rule.pattern = pattern
		} else if pattern, ok = strings.CutSuffix(rule.raw, "*"); ok {
			result.prefixes[rule.raw] = rule
			rule.pattern = pattern
		} else {
			result.endpoints[rule.raw] = rule
			rule.pattern = rule.raw // exact match
		}
	}

	return result
}

func (s *StarRulePathResolver) IsPathMatching(path string) bool {
	if rule, found := s.endpoints[path]; found {
		rule.used = true

		return s.pathMatchingCallback(true)
	}

	for _, prefix := range s.prefixes {
		if strings.HasPrefix(path, prefix.pattern) {
			prefix.used = true

			return s.pathMatchingCallback(true)
		}
	}

	for _, suffix := range s.suffixes {
		if strings.HasSuffix(path, suffix.pattern) {
			suffix.used = true

			return s.pathMatchingCallback(true)
		}
	}

	return s.pathMatchingCallback(false)
}

// NoIDPath ignores URLs that have any ID.
type NoIDPath struct{}

func (NoIDPath) IsPathMatching(path string) bool {
	// There should be no slashes, curly brackets.
	// If there are we are dealing with nester resources. Those are to be ignored.
	return !strings.Contains(path, "{")
}

// NoNestedIDPath allows exactly one ID in URL.
type NoNestedIDPath struct{}

func (NoNestedIDPath) IsPathMatching(path string) bool {
	// Remove the last URL part.
	parts := strings.Split(path, "/")
	parts = parts[:len(parts)-1]
	path = strings.Join(parts, "/")

	// We get the large prefix which may have variate parts. It shouldn't.
	return !strings.Contains(path, "{")
}

type CustomPathMatcher func(path string) bool

func (m CustomPathMatcher) IsPathMatching(path string) bool {
	return m(path)
}
