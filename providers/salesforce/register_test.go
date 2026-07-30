package salesforce

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

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
			name:    "custom object keeps namespace-style separator before ChangeEvent",
			objName: "my_custom_object__c",
			want:    "my_custom_object__ChangeEvent",
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
