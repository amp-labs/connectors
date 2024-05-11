package mockutils

import (
	"fmt"
	"net/http"
)

func RespondWithMethodExpectation(w http.ResponseWriter, r *http.Request, methodName string) {
	// if method is not as expected we return error code so the test will fail
	// and with response payload which will be a helpful message for debugging
	if r.Method == methodName {
		w.WriteHeader(http.StatusNoContent)
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
