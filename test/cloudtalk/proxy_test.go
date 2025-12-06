package cloudtalk

import (
	"context"
	"net/http"
	"testing"
)

func TestConnector_Proxy(t *testing.T) {
	conn := GetCloudTalkConnector(context.Background())

	// CloudTalk Base URL: https://my.cloudtalk.io/api
	// We will use the countries endpoint which we know works without specific IDs
	url := "https://my.cloudtalk.io/api/countries/index.json"

	client := conn.HTTPClient()
	res, _, err := client.Get(context.Background(), url)
	if err != nil {
		t.Fatalf("Client.Get failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}
}
