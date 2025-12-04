package salesforce

import (
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"gotest.tools/v3/assert"
)

func TestSoqlBuilderWithIDs(t *testing.T) {
	t.Parallel()

	soql := makeSOQL(common.ReadParams{
		ObjectName: "Account",
		// Note: fields doesn't preserve order of elements.
		// To simplify test only one element is included.
		Fields: datautils.NewSet("shippingstreet"),
	})

	{
		// SOQL builder must produce query matching documentation.
		// nolint:lll
		// https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select_fields.htm
		soql.WithIDs(
			[]string{
				"001ak00000OQ4RxAAL",
				"001ak00000OQ4RyAAL",
				"001ak00000OQ4TZAA1",
				"001ak00000OQ4TbAAL",
				"001ak00000OQ4VBAA1",
				"001ak00000OQ4VCAA1",
			})

		output := soql.String()
		assert.Equal(t, output, "SELECT shippingstreet FROM Account WHERE Id IN ("+
			"'001ak00000OQ4RxAAL',"+
			"'001ak00000OQ4RyAAL',"+
			"'001ak00000OQ4TZAA1',"+
			"'001ak00000OQ4TbAAL',"+
			"'001ak00000OQ4VBAA1',"+
			"'001ak00000OQ4VCAA1')", "mismatching SOQL query string")
	}
}

func TestSoqlBuilderWithParentAssociation(t *testing.T) {
	t.Parallel()

	// Test that AccountId is added to SOQL when account is requested as association
	soql := makeSOQL(common.ReadParams{
		ObjectName:        "opportunity",
		Fields:            datautils.NewSet("Name", "Amount"),
		AssociatedObjects: []string{"account"},
	})

	output := soql.String()
	// AccountId should be included in the SELECT clause
	// Note: field order may vary, so we just check for presence
	assert.Assert(t, containsFieldInSOQL(output, "AccountId"),
		"AccountId should be in SOQL query when account is requested as association")
	assert.Assert(t, containsFieldInSOQL(output, "Name"), "Name should be in SOQL query")
	assert.Assert(t, containsFieldInSOQL(output, "Amount"), "Amount should be in SOQL query")
	assert.Assert(t, strings.Contains(output, "FROM opportunity"), "FROM clause should be present")
}

func TestSoqlBuilderWithJunctionAssociation(t *testing.T) {
	t.Parallel()

	// Test that OpportunityContactRoles subquery is added to SOQL when contacts is requested
	// as association for Opportunity
	soql := makeSOQL(common.ReadParams{
		ObjectName:        "opportunity",
		Fields:            datautils.NewSet("Name", "Amount"),
		AssociatedObjects: []string{"contacts"},
	})

	output := soql.String()
	// OpportunityContactRoles subquery should be included in the SELECT clause
	assert.Assert(t, strings.Contains(output, "(SELECT FIELDS(STANDARD) FROM OpportunityContactRoles)"),
		"OpportunityContactRoles subquery should be in SOQL query when contacts is requested as association for Opportunity")
	assert.Assert(t, containsFieldInSOQL(output, "Name"), "Name should be in SOQL query")
	assert.Assert(t, containsFieldInSOQL(output, "Amount"), "Amount should be in SOQL query")
	assert.Assert(t, strings.Contains(output, "FROM opportunity"), "FROM clause should be present")
}

// containsFieldInSOQL checks if a field name appears in the SOQL SELECT clause.
func containsFieldInSOQL(soql, fieldName string) bool {
	// Simple check: look for the field name in the SELECT clause
	// This is a basic implementation - in production you might want more robust parsing
	selectStart := 0

	if idx := findString(soql, "SELECT "); idx != -1 {
		selectStart = idx + 7 // len("SELECT ")
	}

	selectEnd := findString(soql[selectStart:], " FROM")

	if selectEnd == -1 {
		return false
	}

	selectClause := soql[selectStart : selectStart+selectEnd]

	return findString(selectClause, fieldName) != -1
}

func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}
