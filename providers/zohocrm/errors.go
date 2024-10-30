package zohocrm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

/*
doc: https://www.zoho.com/crm/developer/docs/api/v6/get-records.html

The normal standard response from ZohoCRM when Reading Records look like this:
{
	"data": [{...}, {...}],
	"info": {...}
}
*/

func responseHandler(resp *http.Response) (*http.Response, error) { //nolint:cyclop
	// When there is no new record after the specified time `since`, ZohoCRM returns `304 Status Not Modified`.
	// Then we wrap this response to 200 Status Okay with empty array data.
	if resp.StatusCode == http.StatusNotModified {
		// Build an empty Response Result (Mimicking ZohoResponse with empty data)
		responseData := map[string]any{
			"data": []any{},
			"info": map[string]any{
				"page":         1,
				"more_records": false,
			},
		}

		data, err := json.Marshal(responseData)
		if err != nil {
			return nil, err
		}

		resp.StatusCode = http.StatusOK
		resp.Body = io.NopCloser(bytes.NewBuffer(data))
	}

	return resp, nil
}
