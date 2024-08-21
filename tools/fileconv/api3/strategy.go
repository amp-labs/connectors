package api3

import (
	"strings"

	"github.com/amp-labs/connectors/common/handy"
)

type ignorePathStrategy struct {
	ignoreEndpoints handy.Set[string]
	prefixes        []string
	suffixes        []string
}

func newIgnorePathStrategy(endpoints []string) *ignorePathStrategy {
	result := &ignorePathStrategy{
		ignoreEndpoints: handy.NewSet([]string{}),
		prefixes:        make([]string, 0),
		suffixes:        make([]string, 0),
	}

	for _, endpoint := range endpoints {
		if rule, ok := strings.CutPrefix(endpoint, "*"); ok {
			result.suffixes = append(result.suffixes, rule)
		} else if rule, ok = strings.CutSuffix(endpoint, "*"); ok {
			result.prefixes = append(result.prefixes, rule)
		} else {
			result.ignoreEndpoints.AddOne(endpoint)
		}
	}

	return result
}

// Check will return true if URL path should be ignored.
func (s ignorePathStrategy) Check(path string) bool {
	if s.ignoreEndpoints.Has(path) {
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
