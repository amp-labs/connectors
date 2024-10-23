package zohocrm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

/*
The normal standard response from ZohoCRM look like this:
{
	"data": [{...}, {...}],
	"info": {...}
}
*/

// responseHandler wraps the http.StatusNotModified response into http.StatusOK
// and sends back an empty data field response.
func responseHandler(resp *http.Response) (*http.Response, error) { //nolint:cyclop
	// responseData represents data we will send back when the response status code 304.
	responseData := map[string]any{
		"data": []any{},
		"info": map[string]any{
			"page":         1,
			"more_records": false,
		},
	}

	// When the ZohoCRM API responds with 304 status code,
	// this indicates there is no data modified since the modification time provided.
	// We modify the response to return 200, And an empty data field response.
	if resp.StatusCode == http.StatusNotModified {
		// Build an empty Response Result (Mimicking ZohoResponse with empty data)
		data, err := json.Marshal(responseData)
		if err != nil {
			return nil, err
		}

		resp.StatusCode = http.StatusOK
		resp.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	return resp, nil
}
