package odoo

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

// E.g POST /json/2/ir.model/search_read.
const jsonAPIVersion = "2"

var (
	errNoIrModelRow   = errors.New("odoo: no ir.model row for model")
	errNoFieldRows    = errors.New("odoo: ir.model.fields search_read returned no rows")
	errSearchReadBody = errors.New("odoo: could not parse search_read response body")
)

type irModelRecord struct {
	Name string `json:"name"`
}

type irModelFieldRow struct {
	Name     string `json:"name"`
	Ttype    string `json:"ttype"`
	Required *bool  `json:"required"`
	Readonly *bool  `json:"readonly"`
	// Relation: if there is no relation Odoo sends a boolean (false), not a string.
	Relation any    `json:"relation"`
	State    string `json:"state"`
}

// listObjectMetadata loads object display names from ir.model and field labels from
// ir.model.fields. The two Odoo calls run in parallel for each object name.
func (c *Connector) listObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		meta, err := c.fetchObjectMetadata(ctx, objectName)
		if err != nil {
			result.AppendError(objectName, err)

			continue
		}

		result.Result[objectName] = *meta
	}

	return result, nil
}

func (c *Connector) fetchObjectMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	var (
		displayName string
		fields      common.FieldsMetadata
	)

	callbacks := []simultaneously.Job{
		func(ctx context.Context) error {
			dn, err := c.searchReadIrModelDisplayName(ctx, objectName)
			if err != nil {
				return err
			}

			displayName = dn

			return nil
		},
		func(ctx context.Context) error {
			fm, err := c.searchReadIrModelFields(ctx, objectName)
			if err != nil {
				return err
			}

			fields = fm

			return nil
		},
	}

	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		return nil, err
	}

	return common.NewObjectMetadata(displayName, fields), nil
}

// get display name from ir.model
// E.g POST /json/2/ir.model/search_read
// Body:
//
//	{
//		"domain": [["model", "=", "crm.lead"]],
//		"fields": ["name", "model"],
//		"limit": false
//	}
//
// Response:
// [
//
//	{
//		"id": 4177,
//		"name": "Lead",
//		"model": "crm.lead"
//	},
//	...
//
// ].
func (c *Connector) searchReadIrModelDisplayName(ctx context.Context, model string) (string, error) {
	urlStr, err := c.getURL("ir.model", "search_read")
	if err != nil {
		return "", err
	}

	reqBody := map[string]any{
		"domain": []any{[]any{"model", "=", model}},
		"fields": []string{"name", "model"},
		"limit":  false,
	}

	resp, err := c.JSONHTTPClient().Post(ctx, urlStr, reqBody)
	if err != nil {
		return "", err
	}

	rows, err := common.UnmarshalJSON[[]irModelRecord](resp)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errSearchReadBody, err)
	}

	if rows == nil || len(*rows) == 0 {
		return "", fmt.Errorf("model %q: %w", model, errNoIrModelRow)
	}

	name := (*rows)[0].Name
	if name == "" {
		return model, nil
	}

	return name, nil
}

// get fields metadata from ir.model.fields
// E.g POST /json/2/ir.model.fields/search_read
// Body:
//
//	{
//		"domain": [["model", "=", "crm.lead"]],
//		"fields": [
//				  "name",
//				  "ttype",
//				  "required",
//				  "readonly",
//				  "relation",
//				  "state"
//		]
//	  }
//
// Response:
// [
//
//	{
//		"id": 4177,
//		"name": "write_date",
//		"ttype": "datetime",
//		"required": false,
//		"readonly": true,
//		"relation": false,
//		"state": "base"
//	},
//
//	...
//
// ].
func (c *Connector) searchReadIrModelFields(ctx context.Context, model string) (common.FieldsMetadata, error) {
	urlStr, err := c.getURL("ir.model.fields", "search_read")
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"domain": []any{[]any{"model", "=", model}},
		"fields": []string{
			"name",
			"ttype",
			"required",
			"readonly",
			"relation",
			"state",
		},
		"limit": false,
	}

	resp, err := c.JSONHTTPClient().Post(ctx, urlStr, body)
	if err != nil {
		return nil, err
	}

	rows, err := common.UnmarshalJSON[[]irModelFieldRow](resp)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSearchReadBody, err)
	}

	if rows == nil || len(*rows) == 0 {
		return nil, fmt.Errorf("model %q: %w", model, errNoFieldRows)
	}

	out := make(common.FieldsMetadata)

	for _, row := range *rows {
		out[row.Name] = fieldMetadataFromIrModelFieldRow(row)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("model %q: %w", model, errNoFieldRows)
	}

	return out, nil
}

// fieldMetadataFromIrModelFieldRow maps ir.model.fields columns into FieldMetadata (typed search_read row).
func fieldMetadataFromIrModelFieldRow(row irModelFieldRow) common.FieldMetadata {
	meta := common.FieldMetadata{
		DisplayName:  row.Name,
		ValueType:    odooTypeToValueType(row.Ttype),
		ProviderType: row.Ttype,
		ReadOnly:     row.Readonly,
		IsRequired:   row.Required,
		IsCustom:     isCustomFromOdooState(row.State),
	}

	rel := relationModel(row.Relation)
	if odooRelationalTType(row.Ttype) && rel != "" {
		meta.ReferenceTo = []string{rel}
	}

	return meta
}
