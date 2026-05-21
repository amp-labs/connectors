package connector

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

// Test_extractAlternateTimestampUsingObjects covers the helper that parses
// per-object opt-in metadata keys ("<ObjectName>_timestampColumnCreateTime")
// into the set passed to salesforce.WithTimestampColumn. The runtime lookup
// in salesforce.Connector.getTimestampColumn lowercases the object name, so
// the map keys must also be lowercase.
func Test_extractAlternateTimestampUsingObjects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata map[string]string
		want     map[common.ObjectName]bool
	}{
		{
			name:     "nil metadata yields empty set",
			metadata: nil,
			want:     map[common.ObjectName]bool{},
		},
		{
			name:     "empty metadata yields empty set",
			metadata: map[string]string{},
			want:     map[common.ObjectName]bool{},
		},
		{
			name: "timestampColumn alone (no per-object keys) yields empty set",
			metadata: map[string]string{
				"timestampColumn": "MyTimestamp__c",
			},
			want: map[common.ObjectName]bool{},
		},
		{
			name: "single opt-in stored lowercase",
			metadata: map[string]string{
				"Account_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{
				"account": true,
			},
		},
		{
			name: "multiple opt-ins all stored lowercase",
			metadata: map[string]string{
				"Account_timestampColumnCreateTime": "true",
				"Contact_timestampColumnCreateTime": "true",
				"Lead_timestampColumnCreateTime":    "true",
			},
			want: map[common.ObjectName]bool{
				"account": true,
				"contact": true,
				"lead":    true,
			},
		},
		{
			name: "custom object with double underscore is preserved",
			metadata: map[string]string{
				"MyCustom__c_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{
				"mycustom__c": true,
			},
		},
		{
			name: "object name with single underscore is preserved",
			metadata: map[string]string{
				"My_Object_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{
				"my_object": true,
			},
		},
		{
			name: "uppercase object name folded to lowercase",
			metadata: map[string]string{
				"ACCOUNT_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{
				"account": true,
			},
		},
		{
			name: "unrelated keys are ignored",
			metadata: map[string]string{
				"timestampColumn":                   "MyTimestamp__c",
				"businessUnitId":                    "abc-123",
				"isDemo":                            "true",
				"Account_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{
				"account": true,
			},
		},
		{
			name: "suffix not at end of key is ignored",
			metadata: map[string]string{
				"Account_timestampColumnCreateTime_extra": "true",
			},
			want: map[common.ObjectName]bool{},
		},
		{
			name: "bare suffix with empty prefix is ignored",
			metadata: map[string]string{
				"_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{},
		},
		{
			name: "value is irrelevant — only key presence matters",
			metadata: map[string]string{
				"Account_timestampColumnCreateTime": "",
				"Contact_timestampColumnCreateTime": "false",
			},
			want: map[common.ObjectName]bool{
				"account": true,
				"contact": true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractAlternateTimestampUsingObjects(tt.metadata)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractAlternateTimestampUsingObjects() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_newSalesforceConnector_alternateTimestampUsingObjects asserts directly
// on the salesforce.Connector's alternateTimestampUsingObjects set after
// newSalesforceConnector runs. This is the structural counterpart to
// Test_newSalesforceConnector_timestampColumn, which only checks behavior
// via GetTimestampColumn.
func Test_newSalesforceConnector_alternateTimestampUsingObjects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata map[string]string
		want     map[common.ObjectName]bool
	}{
		{
			name:     "no metadata yields empty set",
			metadata: nil,
			want:     map[common.ObjectName]bool{},
		},
		{
			name: "metadata without timestampColumn: per-object keys are NOT consumed",
			metadata: map[string]string{
				"Account_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{},
		},
		{
			name: "empty timestampColumn: per-object keys are NOT consumed",
			metadata: map[string]string{
				"timestampColumn":                   "",
				"Account_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{},
		},
		{
			name: "timestampColumn alone: empty set",
			metadata: map[string]string{
				"timestampColumn": "MyTimestamp__c",
			},
			want: map[common.ObjectName]bool{},
		},
		{
			name: "multiple opt-ins are stored lowercase",
			metadata: map[string]string{
				"timestampColumn":                   "MyTimestamp__c",
				"Account_timestampColumnCreateTime": "true",
				"Contact_timestampColumnCreateTime": "true",
				"Lead_timestampColumnCreateTime":    "true",
			},
			want: map[common.ObjectName]bool{
				"account": true,
				"contact": true,
				"lead":    true,
			},
		},
		{
			name: "custom object with double underscore preserved",
			metadata: map[string]string{
				"timestampColumn":                       "MyTimestamp__c",
				"MyCustom__c_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{
				"mycustom__c": true,
			},
		},
		{
			name: "unrelated keys are ignored",
			metadata: map[string]string{
				"timestampColumn":                   "MyTimestamp__c",
				"businessUnitId":                    "abc-123",
				"isDemo":                            "true",
				"Account_timestampColumnCreateTime": "true",
			},
			want: map[common.ObjectName]bool{
				"account": true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn, err := newSalesforceConnector(common.ConnectorParams{
				AuthenticatedClient: mockutils.NewClient(),
				Workspace:           "test-workspace",
				Module:              providers.ModuleSalesforceCRM,
				Metadata:            tt.metadata,
			})
			if err != nil {
				t.Fatalf("newSalesforceConnector() returned unexpected error: %v", err)
			}

			got := conn.AlternateTimestampUsingObjects()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("connector.alternateTimestampUsingObjects = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_newSalesforceConnector_timestampColumn exercises newSalesforceConnector
// end-to-end and asserts that the per-object opt-in metadata actually lands
// in the Connector's alternateTimestampUsingObjects set, via the exported
// GetTimestampColumn behavioral accessor.
//
// For each case: objects listed in optedIn must return the configured
// alternate column; objects listed in notOptedIn must return the default
// (SystemModstamp). When no override column is supplied, every object falls
// back to the default regardless of metadata.
func Test_newSalesforceConnector_timestampColumn(t *testing.T) {
	t.Parallel()

	const (
		altColumn     = "MyTimestamp__c"
		defaultColumn = "SystemModstamp"
	)

	tests := []struct {
		name       string
		metadata   map[string]string
		optedIn    map[common.ObjectName]string // object → expected column
		notOptedIn []common.ObjectName          // objects that must fall back to default
	}{
		{
			name:       "no metadata: everything uses default",
			metadata:   nil,
			notOptedIn: []common.ObjectName{"Account", "Contact"},
		},
		{
			name: "metadata without timestampColumn: everything uses default",
			metadata: map[string]string{
				"businessUnitId":                    "abc-123",
				"Account_timestampColumnCreateTime": "true",
			},
			notOptedIn: []common.ObjectName{"Account", "Contact"},
		},
		{
			name: "empty timestampColumn is ignored even when opt-ins present",
			metadata: map[string]string{
				"timestampColumn":                   "",
				"Account_timestampColumnCreateTime": "true",
			},
			notOptedIn: []common.ObjectName{"Account", "Contact"},
		},
		{
			name: "timestampColumn set without opt-ins: still default everywhere",
			metadata: map[string]string{
				"timestampColumn": altColumn,
			},
			notOptedIn: []common.ObjectName{"Account", "Contact"},
		},
		{
			name: "opted-in objects get alternate, others get default",
			metadata: map[string]string{
				"timestampColumn":                   altColumn,
				"Account_timestampColumnCreateTime": "true",
				"Contact_timestampColumnCreateTime": "true",
			},
			optedIn: map[common.ObjectName]string{
				"Account": altColumn,
				"Contact": altColumn,
			},
			notOptedIn: []common.ObjectName{"Lead", "Opportunity"},
		},
		{
			name: "object lookup is case-insensitive against metadata key casing",
			metadata: map[string]string{
				"timestampColumn":                   altColumn,
				"Account_timestampColumnCreateTime": "true",
			},
			optedIn: map[common.ObjectName]string{
				"Account": altColumn, // exact
				"account": altColumn, // lowercase
				"ACCOUNT": altColumn, // uppercase
			},
		},
		{
			name: "custom object with double underscore opts in",
			metadata: map[string]string{
				"timestampColumn":                       altColumn,
				"MyCustom__c_timestampColumnCreateTime": "true",
			},
			optedIn: map[common.ObjectName]string{
				"MyCustom__c": altColumn,
			},
			notOptedIn: []common.ObjectName{"Account"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn, err := newSalesforceConnector(common.ConnectorParams{
				AuthenticatedClient: mockutils.NewClient(),
				Workspace:           "test-workspace",
				Module:              providers.ModuleSalesforceCRM,
				Metadata:            tt.metadata,
			})
			if err != nil {
				t.Fatalf("newSalesforceConnector() returned unexpected error: %v", err)
			}

			if conn == nil {
				t.Fatal("newSalesforceConnector() returned nil connector")
			}

			for obj, wantCol := range tt.optedIn {
				if got := conn.GetTimestampColumn(obj); got != wantCol {
					t.Errorf("GetTimestampColumn(%q) = %q, want %q (opted-in object)", obj, got, wantCol)
				}
			}

			for _, obj := range tt.notOptedIn {
				if got := conn.GetTimestampColumn(obj); got != defaultColumn {
					t.Errorf("GetTimestampColumn(%q) = %q, want %q (object not opted in)", obj, got, defaultColumn)
				}
			}
		})
	}
}
