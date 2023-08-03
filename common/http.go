package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func DoHttpGetCall(c *http.Client, baseUrl string, endpoint string, token string) (result map[string]interface{}, e *ErrorWithStatus) {
	url, err := url.JoinPath(baseUrl, endpoint)
	if err != nil {
		return nil, &ErrorWithStatus{
			Mode:    NonApiError,
			Message: fmt.Sprintf("error constructing URL: %v", err),
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &ErrorWithStatus{
			Mode:    NonApiError,
			Message: fmt.Sprintf("error creating request: %v", err),
		}
	}

	req.Header.Add("Authorization", "OAuth "+token)
	res, err := c.Do(req)

	if err != nil {
		return nil, &ErrorWithStatus{
			Mode:    NonApiError,
			Message: fmt.Sprintf("error adding authorization header: %v", err),
		}
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return nil, &ErrorWithStatus{
			Mode:    NonApiError,
			Message: fmt.Sprintf("error reading response body: %v", err),
		}
	}

	if res.StatusCode != 200 {
		return nil, &ErrorWithStatus{
			HttpStatus: res.StatusCode,
			Message:    fmt.Sprintf("original message from API server: %v", string(body)),
		}
	}

	d := make(map[string]interface{})
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, &ErrorWithStatus{Message: fmt.Sprintf("failed to unmarshall response body into JSON: %v", err)}
	}
	return d, nil
}
