package api2

import (
	"encoding/json"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/invopop/yaml"
)

// OpenapiFileManager locates openapi file.
// Allows to read data of interest.
// Use it when dealing with OpenAPI v2.
type OpenapiFileManager struct {
	file []byte
}

func NewOpenapiFileManager(file []byte) *OpenapiFileManager {
	return &OpenapiFileManager{
		file: file,
	}
}

func (m OpenapiFileManager) GetExplorer(opts ...api3.Option) (*api3.Explorer, error) {
	dataV2, err := parseV2(m.file)
	if err != nil {
		return nil, err
	}

	dataV3, err := openapi2conv.ToV3(dataV2)
	if err != nil {
		return nil, err
	}

	return api3.NewExplorer(dataV3, opts...), nil
}

func parseV2(file []byte) (*openapi2.T, error) {
	var data openapi2.T

	if err := yaml.Unmarshal(file, &data); err != nil {
		// YAML parsing failed. Fallback to JSON.
		if err = json.Unmarshal(file, &data); err != nil {
			// Cannot be parsed neither as YAML nor JSON.
			return nil, err
		}
	}

	return &data, nil
}
