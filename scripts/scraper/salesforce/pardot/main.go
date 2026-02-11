// nolint:forbidigo,mnd
package main

import (
	"fmt"
	"log"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/scripts/scraper/salesforce/pardot/internal/files"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// enumExceptions contains hard-coded enum values for fields that cannot be
// reliably scraped from the Account Engagement documentation.
//
// In most cases, enum values appear in the object tables and are extracted
// automatically. Some objects, however, document allowed values in separate
// sections or only in narrative text. Those cases are handled here.
var enumExceptions = datautils.Map[string, staticschema.FieldValues]{ // nolint:gochecknoglobals
	// https://developer.salesforce.com/docs/marketing/pardot/guide/dynamic-content-v5.html
	"dynamic-contents.basedOn": makeList("default", "custom", "grade", "score"),
	// https://developer.salesforce.com/docs/marketing/pardot/guide/dynamic-content-variation.html#enums
	"dynamic-content-variations.operator": makeList("is", "contains", "between", "less-than",
		"less-than-or-equal", "greater-than", "greater-than-or-equal"),
	// https://developer.salesforce.com/docs/marketing/pardot/guide/engagement-studio-program-v5.html#enums
	"engagement-studio-programs.status": makeList("draft", "running", "paused", "starting", "scheduled"),
	// https://developer.salesforce.com/docs/marketing/pardot/guide/landing-page-v5.html#enums
	"landing-pages.layoutType": makeList("Layout Template", "Landing Page Builder",
		"Legacy Page Builder", "Salesforce Builder",
	),
	// https://developer.salesforce.com/docs/marketing/pardot/guide/lifecycle-stage-v5.html#read-only-fields
	"lifecycle-stages.matchType": makeList("all", "any"),
	// https://developer.salesforce.com/docs/marketing/pardot/guide/tagged-object-v5.html#enums
	"tagged-objects.objectType": makeList(
		"campaign", "custom-field", "custom-redirect", "dynamic-content",
		"email", "email-template", "file", "form-field", "form-handler", "form",
		"landing-page", "layout-template", "list", "prospect", "user"),
}

func main() {
	// Index file was created manually.
	createSchemas()

	slog.Info("Completed.")
}

func createSchemas() {
	index, err := files.OutputSalesforcePardot.LoadIndex()
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()

	for identifier, model := range index.ModelDocs {
		doc := scrapper.QueryHTML(model.URL)

		urlPath := extractURL(doc)
		objectName := model.Name
		displayName := model.DisplayName

		fields := slices.Concat(
			extractFieldsFromTable(doc, objectName, "#required-fields", false),
			extractFieldsFromTable(doc, objectName, "#optional-fields", false),
			extractFieldsFromTable(doc, objectName, "#read-only-fields", true),
		)

		if len(fields) == 0 {
			fmt.Println("Object had no fields", objectName, model.URL)
		}

		for _, field := range fields {
			schemas.Add(providers.ModuleSalesforceAccountEngagement,
				objectName, displayName, urlPath, "values",
				field, &model.URL, nil)
		}

		log.Printf("Schemas completed %.2f%% [%v]\n",
			getPercentage(identifier, len(index.ModelDocs)), objectName)
	}

	goutils.MustBeNil(files.OutputSalesforcePardot.FlushSchemas(schemas))
}

func extractFieldsFromTable(doc *goquery.Document,
	objectName, sectionName string, isReadOnly bool,
) []staticschema.FieldMetadataMapV2 {
	section := doc.Find(sectionName).First()

	fields := make([]staticschema.FieldMetadataMapV2, 0)

	section.
		Next().
		Find("tr").
		Each(func(i int, row *goquery.Selection) {
			cols := row.Find("td")
			if cols.Length() < 3 {
				return
			}

			fieldName := cleanText(cols.Eq(0).Text())
			fieldType := cleanText(cols.Eq(1).Text())
			fieldValues := getEnums(objectName, fieldName, fieldType, cols.Eq(2))

			if fieldName == "" {
				return
			}

			if f := createField(objectName, fieldName, fieldType, isReadOnly, fieldValues); f != nil {
				fields = append(fields, f)
			}
		})

	return fields
}

func extractURL(doc *goquery.Document) string {
	supported := doc.Find("#supported-operations").First()

	if supported.Length() == 0 {
		log.Println("supported-operations heading not found")

		return ""
	}

	var queryURL string

	supported.
		Next().
		Find("tr").
		EachWithBreak(func(_ int, row *goquery.Selection) bool {
			cols := row.Find("td")
			if cols.Length() < 4 {
				return true // continue
			}

			op := cleanText(cols.Eq(0).Text())

			if strings.EqualFold(op, "Query") {
				queryURL = cleanText(cols.Eq(2).Text())

				return false // break loop
			}

			return true
		})

	queryURL, _ = strings.CutPrefix(queryURL, "https://pi.pardot.com/api/v5/objects/")
	queryURL, _, _ = strings.Cut(queryURL, "?")

	return queryURL
}

func createField(objectName, fieldName, fieldType string,
	isReadOnly bool, values staticschema.FieldValues,
) staticschema.FieldMetadataMapV2 {
	fieldType = strings.Trim(fieldType, " ")

	fieldVal, ok := getFieldValueType(fieldType)
	if !ok {
		fmt.Printf("Ignoring object field. Object %v, field %v, type %v.\n", objectName, fieldName, fieldType)

		return nil
	}

	return staticschema.FieldMetadataMapV2{
		fieldName: staticschema.FieldMetadata{
			DisplayName:  fieldName,
			ValueType:    fieldVal,
			ProviderType: fieldType,
			ReadOnly:     goutils.Pointer(isReadOnly),
			Values:       values,
		},
	}
}

func getEnums(objectName, fieldName, fieldType string, description *goquery.Selection) staticschema.FieldValues {
	if !strings.EqualFold(fieldType, "Enum") {
		return nil
	}

	// Manually written enums.
	if values, ok := enumExceptions[objectName+"."+fieldName]; ok {
		return values
	}

	var values staticschema.FieldValues

	description.Find("code").Each(func(_ int, c *goquery.Selection) {
		val := cleanText(c.Text())

		val = strings.Trim(val, `"`) // remove quotes
		if val != "" {
			values = append(values, staticschema.FieldValue{
				Value:        val,
				DisplayValue: val,
			})
		}
	})

	if len(values) == 0 {
		fmt.Println("Object field is of enumeration type but has no options", objectName, fieldName)
	}

	return values
}

// https://developer.salesforce.com/docs/marketing/pardot/guide/version5overview.html#data-types
func getFieldValueType(fieldType string) (common.ValueType, bool) {
	switch strings.ToLower(fieldType) {
	case "string":
		return common.ValueTypeString, true
	case "float":
		return common.ValueTypeFloat, true
	case "integer":
		return common.ValueTypeInt, true
	case "boolean":
		return common.ValueTypeBoolean, true
	case "datetime":
		return common.ValueTypeDateTime, true
	case "enum":
		return common.ValueTypeSingleSelect, true
	case "array":
		return common.ValueTypeOther, true
	default:
		// object, array
		return common.ValueTypeOther, false
	}
}

func cleanText(text string) string {
	re := regexp.MustCompile(`^\s+|\s+$`)
	// Replace matches with an empty string
	return re.ReplaceAllString(text, "")
}

func getPercentage(i int, i2 int) float64 {
	return (float64(i+1) / float64(i2)) * 100 // nolint:mnd
}

func makeList(options ...string) staticschema.FieldValues {
	return datautils.ForEach(options, func(option string) staticschema.FieldValue {
		return staticschema.FieldValue{
			Value:        option,
			DisplayValue: option,
		}
	})
}
