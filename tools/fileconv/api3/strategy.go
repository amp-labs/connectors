package api3

import (
	"strings"

	"github.com/amp-labs/connectors/common/handy"
)

type PathMatcher interface {
	IsPathMatching(path string) bool
}

type AllowPathStrategy struct {
	Paths []string
}

func (s AllowPathStrategy) IsPathMatching(path string) bool {
	return newStarRulePathResolver(s.Paths).IsPathMatching(path)
}

type DenyPathStrategy struct {
	Paths []string
}

func (s DenyPathStrategy) IsPathMatching(path string) bool {
	return !newStarRulePathResolver(s.Paths).IsPathMatching(path)
}

// PathMatchingStrategy This resolver has 2 level checking.
// Path matches if it is approved by allow resolver or denied if matches by deny resolver.
// The tie of what takes precedence is resolved by priorityAllow flag.
// If neither match we rely on Strict flag.
// Strict flag will mark a path as ignored, use this when AllowPaths and DenyPaths are exhaustive lists.
// By default, Strict is false meaning any path will match, unless specified.
//
// prunes URL Path that should be omitted or included during Schema extraction.
type PathMatchingStrategy struct {
	AllowPaths    []string
	DenyPaths     []string
	Strict        bool
	PriorityAllow bool
}

func (s PathMatchingStrategy) createResolver() *combinedPathResolver {
	return &combinedPathResolver{
		allow:         newStarRulePathResolver(s.AllowPaths),
		deny:          newStarRulePathResolver(s.DenyPaths),
		strict:        s.Strict,
		priorityAllow: s.PriorityAllow,
	}
}

// Uses two level path matching. One is allowing another is denying paths.
type combinedPathResolver struct {
	allow         PathMatcher
	deny          PathMatcher
	strict        bool
	priorityAllow bool
}

func (c combinedPathResolver) IsPathMatching(path string) bool {
	allow := c.allow.IsPathMatching(path)
	deny := c.deny.IsPathMatching(path)

	if allow && deny {
		// Both rules matched. This is a tie.
		// If "allow priority" overrides "deny rules" then return true, otherwise false.
		return c.priorityAllow
	}

	if !allow && !deny {
		// Path doesn't match any rules.
		// If path definitions are strict then we return false, otherwise all paths are allowed.
		return !c.strict
	}

	if allow {
		return true
	}

	// Reaching this line means that `deny` was true, being the last option.
	// Therefore, path must be ignored.
	return false
}

// This path resolver will report if path matches endpoint rule.
// Match can occur in 3 different ways,
// * exact value is inside the registry
// * or using star rule for prefix, suffix matching.
//
// Ex:
// Basic:	/v1/orders	- matches exact path
// Suffix:	*/batch		- matches paths ending with batch
// Prefix:	/v2/*		- matches paths starting with v2.
type starRulePathResolver struct {
	endpoints handy.StringSet
	prefixes  []string
	suffixes  []string
}

func newStarRulePathResolver(endpoints []string) *starRulePathResolver {
	result := &starRulePathResolver{
		endpoints: handy.NewStringSet(),
		prefixes:  make([]string, 0),
		suffixes:  make([]string, 0),
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
		return true
	}

	for _, prefix := range s.prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	for _, suffix := range s.suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}

	return false
}
