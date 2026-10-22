// nolint:forbidigo
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/internal/codec"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/scraper/wordpress/internal/files"
)

const (
	proxyBaseURL = "http://localhost:4444"
	workspaceID  = "258069112"
)

func main() {
	client := &http.Client{
		Timeout: 30 * time.Second, // nolint:mnd
	}

	responses := make(map[string]json.RawMessage, len(files.Endpoints))
	for _, path := range files.Endpoints {
		url := proxyBaseURL + strings.ReplaceAll(path, "{workspace_id}", workspaceID)

		fmt.Printf("Requesting %s\n", url)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodOptions, url, nil)
		if err != nil {
			fmt.Printf("Create request failed for %s: %v\n", url, err)

			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Request failed for %s: %v\n", url, err)

			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Printf("Read response failed for %s: %v\n", url, err)

			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Printf("Request returned HTTP %d for %s: %s\n", resp.StatusCode, url, string(body))

			continue
		}

		if !json.Valid(body) {
			fmt.Printf("Invalid JSON response from %s\n", url)

			continue
		}

		responses[path] = body
	}

	goutils.MustBeNil(codec.WriteJson(files.CombinedResponsesFileLocation, responses))
	fmt.Printf("Saved %d responses to %s\n", len(responses), files.CombinedResponsesFileLocation)
}
