package scrapper

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// QueryHTML this method can panic.
func QueryHTML(url string) *goquery.Document { //nolint:gocritic
	return QueryLoadableHTML(url, 0)
}

// QueryLoadableHTML this method can panic.
// Uses browser internally to wait for all HTML parts to be fully loaded.
func QueryLoadableHTML(url string, wait int64) *goquery.Document { //nolint:gocritic
	htmlPage, err := loadHTMLPage(url, wait)
	if err != nil {
		log.Fatal(err)
	}

	reader := strings.NewReader(htmlPage)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func loadHTMLPage(url string, wait int64) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Run Chrome headless and extract the final rendered HTML.
	var htmlContent string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// Wait for JavaScript to load content
		chromedp.Sleep(time.Duration(wait)*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)

	return htmlContent, err
}

func LoadFile(filename string, object any) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &object)
}
