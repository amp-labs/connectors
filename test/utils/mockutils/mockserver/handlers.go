package mockserver

import "net/http"

// ContentJSON is a setup handler, which configures server to use JSON.
func ContentJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	}
}

// ContentXML is a setup handler, which configures server to use XML.
func ContentXML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
	}
}

// ContentHTML is a setup handler, which configures server to use HTML.
func ContentHTML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/html")
	}
}

// ContentText is a setup handler, which configures server to use Plain Text.
func ContentText() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
	}
}

// ContentMIME is a setup handler, which configures custom media type.
func ContentMIME(mediaType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", mediaType)
	}
}

// Response is used to configure server response with HTTP status and body data.
// Data is optional.
func Response(status int, data ...[]byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)

		if len(data) == 1 {
			_, _ = w.Write(data[0])
		}

		if len(data) > 1 {
			// The test author made a mistake.
			panic("at most one response body can be returned by mockserver")
		}
	}
}

func ResponseString(status int, data string) http.HandlerFunc {
	return Response(status, []byte(data))
}
