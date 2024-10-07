package mockserver

import "net/http"

// ContentJSON is a setup handler, which configures server to use JSON.
func ContentJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
	}
}

func ResponseString(status int, data string) http.HandlerFunc {
	return Response(status, []byte(data))
}
