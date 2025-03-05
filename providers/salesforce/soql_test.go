package salesforce

import (
	"strconv"
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
		// SOQL query has a limit on how many record identifiers can be queried,
		// exceeding it would produce an error.
		tooManyIdentifiers := make([]string, 210)
		for index := range tooManyIdentifiers {
			tooManyIdentifiers[index] = strconv.Itoa(index)
		}

		err := soql.WithIDs(tooManyIdentifiers)
		assert.ErrorIs(t, err, common.ErrTooManyRecordIDs, "expected to get an error for large ids list")
	}
	{
		// SOQL builder must produce query matching documentation.
		// nolint:lll
		// https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select_fields.htm
		err := soql.WithIDs(
			[]string{
				"001ak00000OQ4RxAAL",
				"001ak00000OQ4RyAAL",
				"001ak00000OQ4TZAA1",
				"001ak00000OQ4TbAAL",
				"001ak00000OQ4VBAA1",
				"001ak00000OQ4VCAA1",
			})
		assert.NilError(t, err, "error should be nil")

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
