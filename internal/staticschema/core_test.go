package staticschema

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestConvertToCommonCarriesReferenceTo(t *testing.T) {
	t.Parallel()

	fields := FieldMetadataMapV2{
		"company": FieldMetadata{
			DisplayName:  "company",
			ValueType:    common.ValueTypeReference,
			ProviderType: "CompanyReference",
			ReferenceTo:  []string{"companies"},
		},
		"firstName": FieldMetadata{
			DisplayName:  "firstName",
			ValueType:    common.ValueTypeString,
			ProviderType: "string",
		},
	}

	result := fields.convertToCommon()

	if got := result["company"].ReferenceTo; !reflect.DeepEqual(got, []string{"companies"}) {
		t.Errorf("company ReferenceTo = %v, want [companies]", got)
	}

	if got := result["firstName"].ReferenceTo; got != nil {
		t.Errorf("firstName ReferenceTo = %v, want nil", got)
	}
}
