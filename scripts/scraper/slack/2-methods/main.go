// nolint:forbidigo
package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

type Method struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Scope struct {
	Name        string   `json:"name"`
	Url         string   `json:"url"`
	Description string   `json:"description"`
	TokenTypes  []string `json:"tokenTypes"`
	Methods     []Method `json:"methods"`
}

func main() {
	var scopes []*Scope
	goutils.MustBeNil(scrapper.LoadFile("scripts/scraper/slack/1-scopes/scopes.json", &scopes))

	for i, scope := range scopes {
		fmt.Printf("%.2f%%\n", float64(i+1)/float64(len(scopes))*100) // nolint:mnd

		scope.Methods = scrapeMethods(scope.Url)
	}

	goutils.MustBeNil(fileconv.Flusher{}.ToFile("scripts/scraper/slack/2-methods/scopes.json", scopes))
}

func scrapeMethods(url string) []Method {
	doc := scrapper.QueryHTML(url)

	var methods []Method

	doc.Find("div.info-row").Each(func(_ int, row *goquery.Selection) {
		key := strings.TrimSpace(row.Find(".info-key").Text())
		if key != "Compatible API methods" {
			return
		}

		row.Find("a").Each(func(_ int, a *goquery.Selection) {
			name := strings.TrimSpace(a.Text())

			href, ok := a.Attr("href")
			if name == "" || !ok || href == "" {
				return
			}

			methods = append(methods, Method{
				Name: name,
				Url:  "https://docs.slack.dev" + href,
			})
		})
	})

	return methods
}
