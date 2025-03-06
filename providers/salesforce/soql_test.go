package salesforce

import (
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
