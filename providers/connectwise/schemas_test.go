package connectwise

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/connectwise/internal/metadata"
)

// TestSchemasReferenceFields pins the reference classification baked into schemas.json
// across objects, so a regeneration that loses lookup detection fails loudly.
// Covers every naming shape ConnectWise uses for reference schemas:
// exact (CompanyReference), prefixed (SystemDepartmentReference),
// namespaced item schema (Finance.Currency -> CurrencyReference)
// and pluralized (BillingTermsReference).
func TestSchemasReferenceFields(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		objectName   string
		fieldName    string
		providerType string
		referenceTo  []string
	}{
		{
			objectName:   "contacts",
			fieldName:    "company",
			providerType: "CompanyReference",
			referenceTo:  []string{"companies"},
		},
		{
			objectName:   "contacts",
			fieldName:    "companyLocation",
			providerType: "SystemLocationReference",
			referenceTo:  []string{"system/locations"},
		},
		{
			objectName:   "sales/opportunities",
			fieldName:    "billingTerms",
			providerType: "BillingTermsReference",
			referenceTo:  []string{"billingTerms"},
		},
		{
			objectName:   "sales/opportunities",
			fieldName:    "contact",
			providerType: "ContactReference",
			referenceTo:  []string{"contacts"},
		},
		{
			objectName:   "sales/opportunities",
			fieldName:    "primarySalesRep",
			providerType: "MemberReference",
			referenceTo:  []string{"system/members"},
		},
		{
			objectName:   "sales/opportunities",
			fieldName:    "currency",
			providerType: "CurrencyReference",
			referenceTo:  []string{"currencies"},
		},
		{
			objectName:   "system/members",
			fieldName:    "defaultDepartment",
			providerType: "SystemDepartmentReference",
			referenceTo:  []string{"system/departments"},
		},
		{
			objectName:   "system/members",
			fieldName:    "reportsTo",
			providerType: "MemberReference",
			referenceTo:  []string{"system/members"},
		},
		{
			// Lookup without a top-level target collection keeps reference type,
			// with no referenceTo.
			objectName:   "contacts",
			fieldName:    "site",
			providerType: "SiteReference",
			referenceTo:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.objectName+"."+tt.fieldName, func(t *testing.T) {
			t.Parallel()

			result, err := metadata.Schemas.Select(common.ModuleRoot, []string{tt.objectName})
			if err != nil {
				t.Fatalf("selecting %v: %v", tt.objectName, err)
			}

			field, ok := result.Result[tt.objectName].Fields[tt.fieldName]
			if !ok {
				t.Fatalf("field %v not found on %v", tt.fieldName, tt.objectName)
			}

			if field.ValueType != common.ValueTypeReference {
				t.Errorf("valueType = %q, want %q", field.ValueType, common.ValueTypeReference)
			}

			if field.ProviderType != tt.providerType {
				t.Errorf("providerType = %q, want %q", field.ProviderType, tt.providerType)
			}

			if !reflect.DeepEqual(field.ReferenceTo, tt.referenceTo) {
				t.Errorf("referenceTo = %v, want %v", field.ReferenceTo, tt.referenceTo)
			}
		})
	}
}
