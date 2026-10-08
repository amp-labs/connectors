// nolint:forbidigo
package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

type Scope struct {
	Name        string   `json:"name"`
	Url         string   `json:"url"`
	Description string   `json:"description"`
	TokenTypes  []string `json:"tokenTypes"`
}

func main() {
	url := "https://docs.slack.dev/reference/scopes/"

	doc := scrapper.QueryLoadableHTML(url, 5) // nolint:mnd

	var scopes []Scope

	doc.Find("div.reference-facts-item").Each(func(_ int, item *goquery.Selection) {
		nameSel := item.Find(".reference-name a")
		name := strings.TrimSpace(nameSel.Text())
		href, _ := nameSel.Attr("href")

		desc := strings.TrimSpace(item.Find(".reference-description").Text())

		var tokenTypes []string
		item.Find(".reference-last-column .reference-subitems-bubble").Each(func(_ int, bubble *goquery.Selection) {
			txt := strings.TrimSpace(bubble.Text())
			if txt != "" {
				tokenTypes = append(tokenTypes, txt)
			}
		})

		if name == "" {
			return
		}

		scopes = append(scopes, Scope{
			Name:        name,
			Url:         "https://docs.slack.dev" + href,
			Description: desc,
			TokenTypes:  tokenTypes,
		})
	})

	goutils.MustBeNil(fileconv.Flusher{}.ToFile("scripts/scraper/slack/1-scopes/scopes.json", scopes))
}
