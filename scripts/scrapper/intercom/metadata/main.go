package main

import (
	"log"
	"log/slog"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/intercom/metadata"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/gertd/go-pluralize"
)

const (
	// IntercomSchemaDocsPrefixURL - root docs URL under which all Models are situated.
	IntercomSchemaDocsPrefixURL = "https://developers.intercom.com"
	// ModelIndexURL - on this page all links to dedicated Models can be found
	// There is no top most URL. Using one of the models situates an index on the left panel.
	ModelIndexURL = "https://developers.intercom.com/docs/references/rest-api/api.intercom.io/Models/activity_log/"
)

// TODO this script got outdated with recent UI change.
func main() {
	// index will have URLs for every schema
	createIndex()
	// using index file collect response fields for every object
	createSchemas()

	slog.Info("Completed.")
}

func createIndex() {
	doc := scrapper.QueryHTML(ModelIndexURL)

	sections := make([]string, 0)

	doc.Find(
		`.MenuGroup__MenuWrapper-sc-qcwcjb-1 div[data-component-name="Menu/MenuGroup"]`).
		Each(func(i int, s *goquery.Selection) {
			cell := s.Find("a")
			relativeURL, _ := cell.Attr("href")
			sections = append(sections, relativeURL)
		})

	registry := scrapper.NewModelURLRegistry()

	for i, section := range sections { // nolint:varnamelen
		doc = scrapper.QueryHTML(IntercomSchemaDocsPrefixURL + section)

		doc.Find(`span[data-component-name="Sidebar/MenuLinkItem"]`).Each(func(i int, s *goquery.Selection) {
			cell := s.Find("a")

			fullname := cell.Find(`li[data-component-name="Menu/MenuItemLabel"]`).Text()
			if strings.HasSuffix(fullname, "schema") {
				// Only add schemas. LineItem must end with schema keyword
				name, _ := strings.CutSuffix(fullname, "schema")
				relativeURL, _ := cell.Attr("href")
				registry.Add(name, IntercomSchemaDocsPrefixURL+relativeURL)
			}
		})

		log.Printf("Index completed %.2f%%\n", getPercentage(i, len(sections))) // nolint:forbidigo
	}

	must(metadata.FileManager.SaveIndex(registry))
}

func createSchemas() {
	index, err := metadata.FileManager.LoadIndex()
	must(err)

	schemas := scrapper.NewObjectMetadataResult()

	// Not all schemas are important.
	// Single GET methods to be discarded. Only match schemas that describe single item for LIST endpoint.
	documents := getSchemasForListEndpoints(index)

	for i := range documents {
		model := documents[i]
		doc := scrapper.QueryHTML(model.URL)

		doc.Find(`.field-name`).Each(func(i int, s *goquery.Selection) {
			name := s.Text()
			schemas.Add(model.Name, model.DisplayName, name, &model.URL)
		})

		log.Printf("Schemas completed %.2f%% [%v]\n", getPercentage(i, len(documents)), model.Name)
	}

	must(metadata.FileManager.SaveSchemas(schemas))
}

func getSchemasForListEndpoints(index *scrapper.ModelURLRegistry) scrapper.ModelDocLinks {
	listSchemas := make(scrapper.ModelDocLinks, 0)

	for _, doc := range index.ModelDocs {
		if schemaName, isList := strings.CutSuffix(doc.Name, "_list"); isList {
			if schema, found := index.ModelDocs.FindByName(schemaName); found {
				listSchemas = append(listSchemas, adaptSchemaToMatchObjectName(schema))
			} else {
				log.Printf("There is no schema describing list elements [%v]\n", schemaName)
			}
		}
	}

	return listSchemas
}

func adaptSchemaToMatchObjectName(schema scrapper.ModelDocLink) scrapper.ModelDocLink {
	// ObjectNames must be plural.
	return scrapper.ModelDocLink{
		Name:        pluralize.NewClient().Plural(schema.Name),
		DisplayName: pluralize.NewClient().Plural(schema.DisplayName),
		URL:         schema.URL,
	}
}

func getPercentage(i int, i2 int) float64 {
	return (float64(i+1) / float64(i2)) * 100 // nolint:gomnd
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
