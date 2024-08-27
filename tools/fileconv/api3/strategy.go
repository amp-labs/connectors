package api3

import (
	"strings"

	"github.com/amp-labs/connectors/common/handy"
)

// This strategy prunes URL Path that should be omitted during Schema extraction.
// It allows hard coded endpoint Path, as well as simple star rules focusing on matching prefix or suffix.
// Ex:
// Basic:	/v1/orders	- ignores this path
// Suffix:	*/batch		- ignores paths ending with batch
// Prefix:	/v2/*		- ignores paths starting with v2.
type ignorePathStrategy struct {
	ignoreEndpoints handy.StringSet
	prefixes        []string
	suffixes        []string
}

func newIgnorePathStrategy(endpoints []string) *ignorePathStrategy {
	result := &ignorePathStrategy{
		ignoreEndpoints: handy.NewStringSet(),
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
