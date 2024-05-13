package mockutils

import (
	"fmt"
	"net/http"
)

func RespondNoContentForMethod(w http.ResponseWriter, r *http.Request, methodName string) {
	RespondToMethod(w, r, methodName, func() {
		w.WriteHeader(http.StatusNoContent)
	})
}

func RespondToMethod(w http.ResponseWriter, r *http.Request, methodName string, onSuccess func()) {
	// if method is not as expected we return error code so the test will fail
	// and with response payload which will be a helpful message for debugging
	if r.Method == methodName {
		// if method is matching execute callback
		onSuccess()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		WriteBody(w, fmt.Sprintf(`{
			"error": {
				"code": "from test",
				"message":"test server expected %v request"
			}}`, methodName))
	}
}

func WriteBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}
