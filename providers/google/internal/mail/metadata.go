package mail

import (
	"context"
	_ "embed"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/scrapper"
)

// nolint:gochecknoglobals
var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	// Schemas is cached data.
	Schemas = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas).MustLoadSchemas()
)

func (a *Adapter) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	result := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, name := range objects {
		nativeObjectName := name
		if objectGroupTypeMessages.Has(name) {
			nativeObjectName = objectNameMessages
		}

		objectMetadata, err := Schemas.SelectOne(providers.ModuleGoogleGmail, nativeObjectName)
		if err == nil {
			switch name {
			case virtualObjectSentMessages:
				objectMetadata.DisplayName = "Sent Messages"
			case virtualObjectInboxMessages:
				objectMetadata.DisplayName = "Messages"
			}

			result.Result[name] = *objectMetadata
		} else {
			result.Errors[name] = err
		}
	}

	return result, nil
}
