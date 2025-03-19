package scrapper

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type query struct{}

var Query = query{} // nolint:gochecknoglobals

func (query) IsVisible(selection *goquery.Selection) bool {
	// Check if the element itself has display: none or visibility: hidden
	style, exists := selection.Attr("style")
	if exists && (strings.Contains(style, "display: none") || strings.Contains(style, "visibility: hidden")) {
		return false
	}

	// Traverse up the parent elements to check if any ancestor is hidden
	parent := selection.Parent()
	for parent.Length() > 0 {
		style, exists = parent.Attr("style")
		if exists && (strings.Contains(style, "display: none") || strings.Contains(style, "visibility: hidden")) {
			return false
		}

		parent = parent.Parent()
	}

	return true
}
