package salesforce

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestIsCustomObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		objName string
		want    bool
	}{
		{name: "standard object", objName: "Account", want: false},
		{name: "custom object", objName: "my_custom_object__c", want: true},
		{name: "camel case custom object", objName: "MyCustomObj__c", want: true},
		{name: "namespaced managed package object", objName: "ns__Employee__c", want: true},
		{name: "custom platform event is not a custom object", objName: "My_Event__e", want: false},
		{name: "custom metadata type is not a custom object", objName: "My_Type__mdt", want: false},
		{name: "name ending in c without suffix separator", objName: "Sync", want: false},
		{name: "empty string", objName: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, IsCustomObject(tt.objName), tt.want)
		})
	}
}

func TestGetChangeDataCaptureEventName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		objName string
		want    string
	}{
		{
			name:    "standard object",
			objName: "Account",
			want:    "AccountChangeEvent",
		},
		{
			name:    "standard object with camel case",
			objName: "OpportunityLineItem",
			want:    "OpportunityLineItemChangeEvent",
		},
		{
			name:    "custom object keeps namespace-style separator before ChangeEvent",
			objName: "my_custom_object__c",
			want:    "my_custom_object__ChangeEvent",
		},
		{
			name:    "camel case custom object",
			objName: "MyCustomObj__c",
			want:    "MyCustomObj__ChangeEvent",
		},
		{
			name:    "namespaced managed package custom object",
			objName: "ns__Employee__c",
			want:    "ns__Employee__ChangeEvent",
		},
		{
			name:    "custom object with digits",
			objName: "Order2Cash__c",
			want:    "Order2Cash__ChangeEvent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, GetChangeDataCaptureEventName(tt.objName), tt.want)
		})
	}
}

func TestGetChangeDataCaptureChannelMembershipName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rawChannelName string
		eventName      string
		want           string
	}{
		{
			name:           "standard object change event",
			rawChannelName: "amp_7d9f0dbc6d3c465ea0515de048981fbc",
			eventName:      "AccountChangeEvent",
			want:           "amp_7d9f0dbc6d3c465ea0515de048981fbc_chn_AccountChangeEvent",
		},
		{
			name:           "standard object with camel case",
			rawChannelName: "amp_x",
			eventName:      "OpportunityLineItemChangeEvent",
			want:           "amp_x_chn_OpportunityLineItemChangeEvent",
		},
		{
			// Salesforce reserves "__" in developer names as the namespace prefix
			// separator. A member name containing my_custom_object__ChangeEvent is
			// parsed as namespace "..._chn_my_custom_object" + name "ChangeEvent"
			// and rejected, so the "__" must be collapsed.
			name:           "custom object change event collapses double underscore",
			rawChannelName: "amp_7d9f0dbc6d3c465ea0515de048981fbc",
			eventName:      "my_custom_object__ChangeEvent",
			want:           "amp_7d9f0dbc6d3c465ea0515de048981fbc_chn_my_custom_object_ChangeEvent",
		},
		{
			// The documented example: SalesEvents_chn_MyCustomObj_ChangeEvent.
			// https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_platformeventchannelmember.htm
			name:           "salesforce docs example for a custom object",
			rawChannelName: "SalesEvents",
			eventName:      "MyCustomObj__ChangeEvent",
			want:           "SalesEvents_chn_MyCustomObj_ChangeEvent",
		},
		{
			name:           "namespaced managed package custom object collapses all double underscores",
			rawChannelName: "amp_x",
			eventName:      "ns__Employee__ChangeEvent",
			want:           "amp_x_chn_ns_Employee_ChangeEvent",
		},
		{
			name:           "custom platform event",
			rawChannelName: "amp_x",
			eventName:      "My_Event__e",
			want:           "amp_x_chn_My_Event_e",
		},
		{
			name:           "custom object change event derived from object name",
			rawChannelName: "amp_x",
			eventName:      GetChangeDataCaptureEventName("my_custom_object__c"),
			want:           "amp_x_chn_my_custom_object_ChangeEvent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetChangeDataCaptureChannelMembershipName(tt.rawChannelName, tt.eventName)
			assert.Equal(t, got, tt.want)
			assert.Assert(t, !strings.Contains(got, "__"),
				"developer names must not contain consecutive underscores: %s", got)
		})
	}
}

func TestGetChannelName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rawChannelName string
		want           string
	}{
		{name: "appends chn suffix", rawChannelName: "amp_x", want: "amp_x__chn"},
		{name: "uuid style channel", rawChannelName: "amp_7d9f0dbc6d3c465ea0515de048981fbc", want: "amp_7d9f0dbc6d3c465ea0515de048981fbc__chn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, GetChannelName(tt.rawChannelName), tt.want)
		})
	}
}

func TestGetRawChannelNameFromChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		channelName string
		want        string
	}{
		{name: "strips chn suffix", channelName: "amp_x__chn", want: "amp_x"},
		{name: "round trips with GetChannelName", channelName: GetChannelName("amp_x"), want: "amp_x"},
		{name: "name without suffix is unchanged", channelName: "amp_x", want: "amp_x"},
		{name: "suffix only", channelName: "__chn", want: ""},
		{name: "empty string", channelName: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, GetRawChannelNameFromChannel(tt.channelName), tt.want)
		})
	}
}

func TestGetRawChannelNameFromObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		objectName string
		want       string
	}{
		{name: "strips platform event suffix", objectName: "My_Event__e", want: "My_Event"},
		{name: "non platform event is unchanged", objectName: "Account", want: "Account"},
		{name: "custom object is unchanged", objectName: "my_custom_object__c", want: "my_custom_object__c"},
		{name: "empty string", objectName: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, GetRawChannelNameFromObject(tt.objectName), tt.want)
		})
	}
}

func TestGetRawObjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		objName string
		want    string
	}{
		{name: "strips custom object suffix", objName: "my_custom_object__c", want: "my_custom_object"},
		{name: "namespaced object keeps namespace prefix", objName: "ns__Employee__c", want: "ns__Employee"},
		{name: "shorter than suffix", objName: "ab", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, GetRawObjectName(tt.objName), tt.want)
		})
	}
}

func TestRemoveSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		objName      string
		suffixLength int
		want         string
	}{
		{name: "removes suffix", objName: "Account__c", suffixLength: 3, want: "Account"},
		{name: "length equal to suffix", objName: "__c", suffixLength: 3, want: ""},
		{name: "shorter than suffix", objName: "ab", suffixLength: 3, want: ""},
		{name: "zero length suffix", objName: "Account", suffixLength: 0, want: "Account"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, RemoveSuffix(tt.objName, tt.suffixLength), tt.want)
		})
	}
}
