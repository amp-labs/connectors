package salesforce

import (
	"fmt"
	"io"
	"net/http"
	"encoding/json"

	"github.com/amp-labs/connectors/common"
)

func makeURL(baseURL string, path string) string {
	// TODO: replace with something better
	return fmt.Sprintf("%s/%s", baseURL, path)
}

func (s SalesforceConnector) MakeGetCall(c common.GetCallConfig) (*common.GenericResult, error) {
	req, error := http.NewRequest("GET", makeURL(s.BaseURL, c.Endpoint), nil)

	if error != nil {
		return nil, error
	}
	req.Header.Add("Authorization", "OAuth "+s.AccessToken)
	response, error := s.Client.Do(req)

	if error != nil {
		return nil, error
	}

	responseBody, error := io.ReadAll(response.Body)

	if error != nil {
		return nil, error
	}

	fmt.Println("Status: ", response.Status)

	d := make(map[string]interface{})
	err := json.Unmarshal(responseBody, &d)
	if err != nil {
		return nil, error
	}
	fmt.Printf("Response body: %v\n\n", d["childRelationships"])
	defer response.Body.Close()
	return &common.GenericResult{Data: d}, nil
}
