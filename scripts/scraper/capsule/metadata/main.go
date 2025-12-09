// nolint:forbidigo
package main

import (
	"fmt"
	"log"
	"log/slog"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/capsule/metadata"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

const (
	Arrow = "→"
	// CapsuleDocsPrefixURL documentation root.
	CapsuleDocsPrefixURL = "https://developer.capsulecrm.com"
	// ModelIndexURL - on this page all links to dedicated Models can be found.
	ModelIndexURL = "https://developer.capsulecrm.com/v2/models/party"
	ModelsPrefix  = "/v2/models/"
)

var (
	ignoreModels = datautils.NewSet( // nolint:gochecknoglobals
		"entry",
		"filter",
		"track",
		"tag",
		"field_definition", // custom fields
		"attachment",
		"currency",
		"site",
		"tag_definition",
	)
	modelNameToObjectName = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
		"activity_types":    "activitytypes",
		"lost_reasons":      "lostreasons",
		"rest_hooks":        "resthooks",
		"track_definitions": "trackdefinitions",
	}, func(objectName string) string {
		return objectName
	})
	displayNameMapping = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
		"activitytypes":    "Activity Types",
		"lostreasons":      "Lost Reasons",
		"resthooks":        "REST Hooks",
		"trackdefinitions": "Track Definitions",
	}, formatDisplay)
	objectNameToURLPath = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
		"parties":          "/v2/parties",
		"opportunities":    "/v2/opportunities",
		"projects":         "/v2/kases",
		"tasks":            "/v2/tasks",
		"users":            "/v2/users",
		"teams":            "/v2/teams",
		"pipelines":        "/v2/pipelines",
		"milestones":       "/v2/milestones",
		"lostreasons":      "/v2/lostreasons",
		"boards":           "/v2/boards",
		"stages":           "/v2/stages",
		"resthooks":        "/v2/resthooks",
		"trackdefinitions": "/v2/trackdefinitions",
		"categories":       "/v2/categories",
		"activitytypes":    "/v2/activitytypes",
		"countries":        "/v2/countries",
		"titles":           "/v2/titles",
	}, func(objectName string) string {
		fmt.Println("No matching URL path for object", objectName)

		return objectName
	})
	objectNameToResponseKey = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
		"parties":          "parties",
		"opportunities":    "opportunities",
		"projects":         "kases",
		"tasks":            "tasks",
		"users":            "users",
		"teams":            "teams",
		"pipelines":        "pipelines",
		"milestones":       "milestones",
		"lostreasons":      "lostReasons",
		"boards":           "boards",
		"stages":           "stages",
		"resthooks":        "restHooks",
		"trackdefinitions": "trackDefinitions",
		"categories":       "categories",
		"activitytypes":    "activityTypes",
		"countries":        "countries",
		"titles":           "personTitles",
	}, func(objectName string) string {
		fmt.Println("No matching response key for object", objectName)

		return objectName
	})
)

func main() {
	createIndex()
	createSchemas()

	slog.Info("Completed.")
}

func createIndex() {
	tabs := getTabLinks()
	registry := scrapper.NewModelURLRegistry()

	for _, tab := range tabs {
		name, _ := strings.CutPrefix(tab, ModelsPrefix)
		if !ignoreModels.Has(name) {
			name = naming.NewPluralString(name).String()
			name = modelNameToObjectName.Get(name)
			registry.AddModelByName(name, CapsuleDocsPrefixURL+tab)
		}
	}

	goutils.MustBeNil(metadata.FileManager.SaveIndex(registry))
}

func createSchemas() {
	index, err := metadata.FileManager.LoadIndex()
	goutils.MustBeNil(err)

	schemas := newScrappedSchemas()

	for i := range index.ModelDocs { // nolint:varnamelen
		model := index.ModelDocs[i]
		doc := scrapper.QueryHTML(model.URL)

		doc.Find(`.object-table tbody tr`).
			Each(func(_ int, tableRow *goquery.Selection) {
				var (
					fieldName   string
					fieldType   string
					specialTag  string
					description string
				)

				// Table has 4 columns.
				// Ex: https://developer.capsulecrm.com/v2/models/user.
				tableRow.Find(`td`).Each(func(columnIndex int, cell *goquery.Selection) {
					text := cleanText(cell.Text())
					// nolint:mnd
					switch columnIndex {
					case 0:
						fieldName = text
					case 1:
						fieldType = text
					case 2:
						specialTag = text
					case 3:
						description = text
					}
				})

				if len(fieldName) == 0 {
					return
				}

				if strings.Contains(specialTag, "Write only") {
					// This field cannot be returned by ListObjectMetadata
					// Omitting field.
					return
				}

				schemas.SaveData(model, fieldName, fieldType, specialTag, description)
			})

		log.Printf("Schemas completed %.2f%% [%v]\n", getPercentage(i, len(index.ModelDocs)), model.Name)
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas.Metadata))
}

func getTabLinks() []string {
	doc := scrapper.QueryHTML(ModelIndexURL)

	links := make([]string, 0)

	doc.Find(".sidebar-section__title").Each(func(i int, s *goquery.Selection) {
		cell := s.Find("a")
		link, _ := cell.Attr("href")

		if strings.HasPrefix(link, ModelsPrefix) {
			links = append(links, link)
		}
	})

	return links
}

func cleanText(text string) string {
	re := regexp.MustCompile(`^\s+|\s+$`)
	// Replace matches with an empty string
	return re.ReplaceAllString(text, "")
}

func implyValueOptions(description string) []string {
	re := regexp.MustCompile(`Accepted values are: ([\w, _]+)`)
	match := re.FindStringSubmatch(description)

	if len(match) == 0 {
		return nil
	}

	return regexp.MustCompile(`\s*,\s*`).Split(match[1], -1)
}

func formatFieldType(fieldType string) string {
	if strings.HasPrefix(fieldType, "Array") {
		return "Array"
	}

	return fieldType
}

func getPercentage(i int, i2 int) float64 {
	return (float64(i+1) / float64(i2)) * 100 // nolint:mnd
}

func formatDisplay(name string) string {
	name = api3.CamelCaseToSpaceSeparated(name)

	return api3.CapitalizeFirstLetterEveryWord(name)
}

func getFieldValueType(fieldType string, fieldOptions []string) common.ValueType {
	switch strings.ToLower(fieldType) {
	case "integer", "long":
		return common.ValueTypeInt
	case "boolean":
		return common.ValueTypeBoolean
	case "date":
		return common.ValueTypeDate
	case "string":
		if len(fieldOptions) != 0 {
			return common.ValueTypeSingleSelect
		}

		return common.ValueTypeString
	default:
		// object, array
		return common.ValueTypeOther
	}
}

func getFieldValueOptions(fieldValueOptions []string) staticschema.FieldValues {
	if len(fieldValueOptions) == 0 {
		return nil
	}

	values := make(staticschema.FieldValues, len(fieldValueOptions))
	for index, option := range fieldValueOptions {
		values[index] = staticschema.FieldValue{
			Value:        option,
			DisplayValue: option,
		}
	}

	return values
}
