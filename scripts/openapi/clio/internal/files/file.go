package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	//go:embed grow.yaml
	growAPI []byte
	//go:embed manage.yaml
	manageAPI []byte

	InputGrow  = api3.NewOpenapiFileManager[any](growAPI)
	OutputGrow = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/clio/internal/grow"),
	)

	InputManage  = api3.NewOpenapiFileManager[any](manageAPI)
	OutputManage = scrapper.NewWriter[staticschema.FieldMetadataMapV2](
		fileconv.NewPath("providers/clio/internal/manage"),
	)
)
