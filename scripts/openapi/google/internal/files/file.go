package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/getkin/kin-openapi/openapi3"
)

// nolint:gochecknoglobals
var (
	// Static file containing openapi spec.
	//go:embed calendar.json
	calendarAPI []byte
	//go:embed people.json
	peopleAPI []byte
	//go:embed mail.json
	mailAPI []byte

	InputCalendar  = api3.NewOpenapiFileManager[any](calendarAPI)
	OutputCalendar = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/google/internal/calendar"))
	InputContacts  = api3.NewOpenapiFileManager[any](peopleAPI)
	OutputContacts = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/google/internal/contacts"))
	InputMail  = api3.NewOpenapiFileManager[any](mailAPI)
	OutputMail = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/google/internal/mail"))
)

// CalendarDocument loads the Calendar OpenAPI spec with all $ref pointers resolved.
// It gives scripts access to component schemas (e.g. nested object properties)
// that the api3 explorer does not surface on its flattened field list.
func CalendarDocument() (*openapi3.T, error) {
	return openapi3.NewLoader().LoadFromData(calendarAPI)
}
