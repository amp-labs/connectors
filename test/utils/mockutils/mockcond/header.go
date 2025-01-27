package mockcond

import (
	"net/http"
)

func HeaderContentURLFormEncoded() Check {
	return Header(http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded"},
	})
}

func Header(header http.Header) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		return headerIsSubset(r.Header, header)
	}
}

func headerIsSubset(superset, subset http.Header) bool {
	for name, values := range subset {
		superValues := make(map[string]bool)
		for _, v := range superset.Values(name) {
			superValues[v] = true
		}
		// every value of this header must be part of superset
		for _, value := range values {
			if _, found := superValues[value]; !found {
				return false
			}
		}
	}

	return true
}
