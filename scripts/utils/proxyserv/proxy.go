package proxyserv

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/future"
)

type Proxy struct {
	*httputil.ReverseProxy

	target *url.URL
}

func newProxy(target *url.URL, httpClient common.AuthenticatedHTTPClient) *Proxy {
	reverseProxy := httputil.NewSingleHostReverseProxy(target)
	reverseProxy.Transport = &customTransport{httpClient}

	return &Proxy{
		ReverseProxy: reverseProxy,
		target:       target,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Host = p.target.Host
	r.Host = p.target.Host
	r.RequestURI = "" // Must be cleared

	fmt.Printf("Proxying request: %s %s%s\n", r.Method, r.URL.Host, r.URL.Path) // nolint:forbidigo
	p.ReverseProxy.ServeHTTP(w, r)
}

func (p *Proxy) Start(ctx context.Context, port int) {
	http.Handle("/", p)

	fmt.Printf("\nProxy server listening on :%d\n", port) // nolint:forbidigo

	if err := listen(ctx, port); err != nil {
		panic(err)
	}
}

type customTransport struct {
	httpClient common.AuthenticatedHTTPClient
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.httpClient.Do(req)
}

// listen will start a server on the given port and block until it is closed.
// This is used as opposed to http.ListenAndServe because it respects the context
// and has a cleaner shutdown sequence.
func listen(ctx context.Context, port int) error {
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 5 * time.Second, // nolint:mnd
	}

	future.GoContext(ctx, func(ctx context.Context) (struct{}, error) {
		<-ctx.Done()

		_ = server.Shutdown(context.Background()) // nolint:contextcheck

		return struct{}{}, nil
	})

	if err := server.Serve(listener); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("HTTP server stopped") // nolint:forbidigo

			return nil
		}

		return err
	}

	return nil
}
