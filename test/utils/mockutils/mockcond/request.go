package mockcond

import (
	"net/http"
	"strings"
)

// PathSuffix returns a check expecting request URL path to match the template.
func PathSuffix(expectedSuffix string) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		path := r.URL.Path

		return strings.HasSuffix(path, expectedSuffix)
	}
}

// Method returns a check expecting HTTP method name to match the template.
func Method(methodName string) Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		return r.Method == methodName
	}
}

func MethodPATCH() Check {
	return Method("PATCH")
}
