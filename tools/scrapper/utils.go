package scrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// QueryHTML this method can panic.
func QueryHTML(url string) *goquery.Document { //nolint:gocritic
	res, err := makeRequest(url)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		_ = res.Body.Close()
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		_ = res.Body.Close()

		log.Fatal(err)
	}

	_ = res.Body.Close()

	return doc
}

func makeRequest(sourceURL string) (*http.Response, error) {
	// must end with `/` to avoid stupid redirect with then without then again with and without slash
	if !strings.HasSuffix(sourceURL, "/") {
		sourceURL += "/"
	}

	client := &http.Client{
		CheckRedirect: func() func(req *http.Request, via []*http.Request) error {
			redirects := 0

			return func(req *http.Request, via []*http.Request) error {
				maxRedirects := 12
				if redirects > maxRedirects {
					return fmt.Errorf("stopped after %v redirects", maxRedirects) // nolint:goerr113
				}
				redirects++

				return nil
			}
		}(),
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func LoadFile(filename string, object any) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &object)
}
