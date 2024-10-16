package api3

import (
	"strings"

	"github.com/amp-labs/connectors/common/handy"
)

// This strategy prunes URL Path that should be omitted or included during Schema extraction.
// It allows hard coded endpoint Path, as well as simple star rules focusing on matching prefix or suffix.
// First mode ignore endpoints and those matching suffixes, prefixes,
// otherwise, only those that match the rules, will be used.
//
// Ex:
// Basic:	/v1/orders	- ignores this path
// Suffix:	*/batch		- ignores paths ending with batch
// Prefix:	/v2/*		- ignores paths starting with v2.
type ignorePathStrategy struct {
	endpoints handy.StringSet
	prefixes  []string
	suffixes  []string
	// ignore flag determines the mode.
	// It changes how we answer questions: Do we "ignore", or do we "include" given endpoint?
	ignore bool
}

func newIgnorePathStrategy(endpoints []string, ignore bool) *ignorePathStrategy {
	result := &ignorePathStrategy{
		endpoints: handy.NewStringSet(),
		prefixes:  make([]string, 0),
		suffixes:  make([]string, 0),
		ignore:    ignore,
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

// Check will return true if URL path should be ignored.
func (s ignorePathStrategy) Check(path string) bool {
	if s.endpoints.Has(path) {
		return s.ignore
	}

	for _, prefix := range s.prefixes {
		if strings.HasPrefix(path, prefix) {
			return s.ignore
		}
	}

	for _, suffix := range s.suffixes {
		if strings.HasSuffix(path, suffix) {
			return s.ignore
		}
	}

	return !s.ignore
}
