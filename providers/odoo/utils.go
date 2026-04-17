package odoo

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

// relationModel maps ir.model.fields "relation" to a model name, or "" when there is no relation
// (Odoo sends boolean false in that case, not an empty string).
func relationModel(v any) string {
	if v == nil {
		return ""
	}

	switch common.InferValueTypeFromData(v) {
	case common.ValueTypeBoolean:
		return ""
	case common.ValueTypeString:
		s, _ := v.(string)

		return strings.TrimSpace(s)
	default:
		return ""
	}
}

// Odoo marks user-defined fields with state "manual"; core fields use "base".
func isCustomFromOdooState(state string) *bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "manual":
		return goutils.Pointer(true)
	case "base":
		return goutils.Pointer(false)
	default:
		return nil
	}
}

func odooRelationalTType(ttype string) bool {
	switch ttype {
	case "many2one", "one2many", "many2many", "reference":
		return true
	default:
		return false
	}
}

func odooTypeToValueType(ttype string) common.ValueType {
	switch ttype {
	case "char", "text", "html":
		return common.ValueTypeString
	case "boolean":
		return common.ValueTypeBoolean
	case "integer":
		return common.ValueTypeInt
	case "float", "monetary":
		return common.ValueTypeFloat
	case "date":
		return common.ValueTypeDate
	case "datetime":
		return common.ValueTypeDateTime
	case "selection":
		return common.ValueTypeSingleSelect
	case "many2one", "one2many", "many2many", "reference":
		return common.ValueTypeReference
	default:
		return common.ValueTypeOther
	}
}
