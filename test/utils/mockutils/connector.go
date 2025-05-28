package mockutils

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

func ReplaceURLOrigin(baseURL, mockServerOrigin string) (alteredBaseURL string) {
	url, err := urlbuilder.New(baseURL)
	if err != nil {
		// Provider URL must follow valid URL format.
		return "impossible"
	}

	return mockServerOrigin + url.Path()
}

// NewClient returns a new instance of an HTTP client.
// Using the shared http.DefaultClient in mock tests may cause CloseIdleConnections errors.
func NewClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}

				return dialer.DialContext(ctx, network, addr)
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
