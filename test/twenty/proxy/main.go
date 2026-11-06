package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/twenty"
	"github.com/amp-labs/connectors/test/utils"
)

// Sends a request the way a proxy caller would: the proxy base URL plus the caller's own path, with the
// connector's authenticated client. go run ./test/twenty/proxy
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetTwentyConnector(ctx)

	config, err := conn.ProxyConfig()
	if err != nil {
		utils.Fail("error reading the proxy config", "error", err)
	}

	for _, path := range []string{"/rest/metadata/objects?limit=1", "/rest/companies?limit=1&depth=0"} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.URL+path, nil)
		if err != nil {
			utils.Fail("error building the request", "error", err)
		}

		resp, err := conn.HTTPClient().Client.Do(req)
		if err != nil {
			utils.Fail("error calling Twenty through the proxy URL", "error", err)
		}

		var body struct {
			TotalCount int `json:"totalCount"`
		}

		_ = json.NewDecoder(resp.Body).Decode(&body)
		_ = resp.Body.Close()

		slog.Info("proxy call", "url", config.URL+path, "status", resp.StatusCode, "totalCount", body.TotalCount)
	}
}
