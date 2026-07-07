package metadata

import (
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestMakeExpandableQueryParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		objectName string
		field      string
		expected   string
	}{
		{
			name:       "empty field",
			objectName: "checkout/sessions",
			field:      "",
			expected:   "",
		},
		{
			name:       "primitive field",
			objectName: "checkout/sessions",
			field:      "id",
			expected:   "",
		},
		{
			name:       "nested checkout session line items",
			objectName: "checkout/sessions",
			field:      "$['line_items']['currency']",
			expected:   "data.line_items",
		},
		{
			name:       "balance transactions top-level source",
			objectName: "balance_transactions",
			field:      "$['source']",
			expected:   "data.source",
		},
		{
			name:       "balance transactions source payment intent",
			objectName: "balance_transactions",
			field:      "$['source']['payment_intent']",
			expected:   "data.source.payment_intent",
		},
		{
			name:       "balance transactions nested path ending in primitive",
			objectName: "balance_transactions",
			field:      "$['source']['payment_intent']['customer']['id']",
			expected:   "data.source.payment_intent.customer",
		},
		{
			name:       "balance transactions nested path ending in expandable object",
			objectName: "balance_transactions",
			field:      "$['source']['payment_intent']['customer']",
			expected:   "data.source.payment_intent.customer",
		},
		{
			name:       "balance transactions depth limit exceeded",
			objectName: "balance_transactions",
			field:      "$['source']['payment_intent']['customer']['discount']",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeExpandableQueryParam(tt.objectName, tt.field)
			res := testutils.NewCompareResult()
			res.Assert("expandable query param", tt.expected, got)
			res.Validate(t, tt.name)
		})
	}
}
