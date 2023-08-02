package salesforce

import (
	"net/http"
	"fmt"
	"io"

	"github.com/amp-labs/connectors/common"
)

func makeURL(baseURL string, path string) string {
	// TODO: replace with something better
	return fmt.Sprintf("%s/%s", baseURL, path)
}

func (s SalesforceConnector) MakeGetCall(c common.GetCallConfig) (*common.GenericResult, error) {
	  request, error := http.NewRequest("GET", makeURL(s.BaseURL, c.Endpoint), nil)

    if error != nil {
        fmt.Println(error)
    }
    response, error := s.Client.Do(request)

    if error != nil {
        fmt.Println(error)
    }
		defer response.Body.Close()	

    responseBody, error := io.ReadAll(response.Body)

    if error != nil {
        fmt.Println(error)
    }

    fmt.Println("Status: ", response.Status)
    fmt.Println("Response body: ", string(responseBody))

		return &common.GenericResult{Data: nil}, nil
}
