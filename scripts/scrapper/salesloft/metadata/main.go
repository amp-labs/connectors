package main

import (
	"flag"
	"log"
	"log/slog"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/salesloft/metadata"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/iancoleman/strcase"
)

const (
	// SalesloftDocsPrefixURL documentation root.
	SalesloftDocsPrefixURL = "https://developers.salesloft.com"
	// ModelIndexURL - on this page all links to dedicated Models can be found.
	ModelIndexURL = "https://developers.salesloft.com/docs/api"
)

var withQueryParamStats bool // nolint:gochecknoglobals

func init() {
	flag.BoolVar(&withQueryParamStats, "queryParamStats", false,
		"collect statistics on query parameters")
	flag.Parse()
}

func main() {
	if withQueryParamStats {
		createQueryParamStats()
	} else {
		// index will have URLs for every schema
		createIndex()
		// using index file collect response fields for every object
		createSchemas()
	}

	slog.Info("Completed.")
}

func createIndex() {
	sections := getSectionLinks()

	registry := scrapper.NewModelURLRegistry()

	for i, section := range sections {
		doc := scrapper.QueryHTML(SalesloftDocsPrefixURL + "/" + section)

		doc.Find(".theme-doc-markdown article").Each(func(i int, s *goquery.Selection) {
			cell := s.Find("a")
			path, _ := cell.Attr("href")
			name, _ := cell.Find("h2").Attr("title")
			registry.Add(name, SalesloftDocsPrefixURL+path)
		})
		log.Printf("Index completed %.2f%%\n", getPercentage(i, len(sections))) // nolint:forbidigo
	}

	must(metadata.FileManager.SaveIndex(registry))
}

func createSchemas() {
	index, err := metadata.FileManager.LoadIndex()
	must(err)

	schemas := scrapper.NewObjectMetadataResult()

	filteredListDocs := getFilteredListDocs(index)
	for i, model := range filteredListDocs { // nolint:varnamelen
		doc := scrapper.QueryHTML(model.URL)

		// There are 2 unordered lists that describe response schema
		modelName := strcase.ToSnake(model.Name)

		doc.Find(`.openapi-tabs__schema-container ul`).
			Each(func(i int, list *goquery.Selection) {
				list.Children().Each(func(i int, property *goquery.Selection) {
					// Sometimes there are nested fields we ignore them
					// Only the first most field represents top level fields of response payload
					fieldName := property.Find(`strong`).First().Text()
					if len(fieldName) != 0 {
						schemas.Add(modelName, model.DisplayName, fieldName)
					}
				})
			})

		log.Printf("Schemas completed %.2f%% [%v]\n", getPercentage(i, len(filteredListDocs)), modelName)
	}

	must(metadata.FileManager.SaveSchemas(schemas))
}

func createQueryParamStats() {
	index, err := metadata.FileManager.LoadIndex()
	must(err)

	registry := make(map[string][]string)

	filteredListDocs := getFilteredListDocs(index)
	numObjects := len(filteredListDocs)

	for i, model := range filteredListDocs { // nolint:varnamelen
		doc := scrapper.QueryHTML(model.URL)

		modelName := strcase.ToSnake(model.Name)

		doc.Find(`.openapi-params__list-item .openapi-schema__property`).Each(func(i int, element *goquery.Selection) {
			prop := element.Text()
			if _, found := registry[prop]; !found {
				registry[prop] = make([]string, 0)
			}

			registry[prop] = append(registry[prop], modelName)
		})

		log.Printf("Query param schemas completed %.2f%% [%v]\n", getPercentage(i, numObjects), modelName)
	}

	// create set of query parameters
	properties := make([]string, 0)

	for prop := range registry {
		if strings.Contains(prop, "[") {
			properties = append(properties, prop)
		}
	}
	// sort query parameters, where most occurred come first
	sort.SliceStable(properties, func(i, j int) bool {
		a := properties[i]
		b := properties[j]
		l1 := len(registry[a])
		l2 := len(registry[b])

		if l1 == l2 {
			return a < b
		}

		return l1 > l2
	})

	// finally prepare to write to file
	stats := scrapper.NewQueryParamStats(numObjects)
	for _, prop := range properties {
		stats.Add(prop, registry[prop])
	}

	must(metadata.FileManager.SaveQueryParamStats(stats))
}

/*
Index file has these suffixes for model name (Total 134):
  - destroy 2
  - call 1 (create-conversations-call, should fall under create category)
  - create 33
  - index 48
  - update 16
  - show 34

`Index` means List operation while `Show` is Singular get.
We are only interested in List schemas (keeping `-index` documents).
*/
func getFilteredListDocs(index *scrapper.ModelURLRegistry) scrapper.ModelDocLinks {
	list := make(scrapper.ModelDocLinks, 0)

	for _, doc := range index.ModelDocs {
		if name, found := strings.CutSuffix(doc.Name, "-index"); found {
			list = append(list, scrapper.ModelDocLink{
				Name:        name,
				DisplayName: doc.DisplayName,
				URL:         doc.URL,
			})
		}
	}

	return list
}

func getSectionLinks() []string {
	doc := scrapper.QueryHTML(ModelIndexURL)

	links := make([]string, 0)

	doc.Find(".margin-top--lg article").Each(func(i int, s *goquery.Selection) {
		cell := s.Find("a")
		link, _ := cell.Attr("href")
		links = append(links, link)
	})

	return links
}

func getPercentage(i int, i2 int) float64 {
	return (float64(i+1) / float64(i2)) * 100 // nolint:gomnd
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
