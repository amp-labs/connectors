package mockutils

import (
	"net/http"
)

func WriteBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}
