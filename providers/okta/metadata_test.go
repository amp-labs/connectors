package okta

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:   "Successfully describe users object",
			Input:  []string{"users"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"id":              "id",
							"status":          "status",
							"created":         "created",
							"activated":       "activated",
							"statusChanged":   "statusChanged",
							"lastLogin":       "lastLogin",
							"lastUpdated":     "lastUpdated",
							"passwordChanged": "passwordChanged",
							"type":            "type",
							"profile":         "profile",
							"credentials":     "credentials",
							"_links":          "_links",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe groups object",
			Input:  []string{"groups"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"groups": {
						DisplayName: "Groups",
						FieldsMap: map[string]string{
							"id":                    "id",
							"created":               "created",
							"lastUpdated":           "lastUpdated",
							"lastMembershipUpdated": "lastMembershipUpdated",
							"objectClass":           "objectClass",
							"type":                  "type",
							"profile":               "profile",
							"_links":                "_links",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe apps object",
			Input:  []string{"apps"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"apps": {
						DisplayName: "Applications",
						FieldsMap: map[string]string{
							"id":          "id",
							"name":        "name",
							"label":       "label",
							"status":      "status",
							"created":     "created",
							"lastUpdated": "lastUpdated",
							"activated":   "activated",
							"signOnMode":  "signOnMode",
							"features":    "features",
							"visibility":  "visibility",
							"credentials": "credentials",
							"settings":    "settings",
							"_links":      "_links",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe logs object",
			Input:  []string{"logs"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"logs": {
						DisplayName: "System Log",
						FieldsMap: map[string]string{
							"uuid":           "uuid",
							"published":      "published",
							"eventType":      "eventType",
							"severity":       "severity",
							"displayMessage": "displayMessage",
							"actor":          "actor",
							"outcome":        "outcome",
							"target":         "target",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe devices object",
			Input:  []string{"devices"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"devices": {
						DisplayName: "Devices",
						FieldsMap: map[string]string{
							"id":           "id",
							"status":       "status",
							"created":      "created",
							"lastUpdated":  "lastUpdated",
							"profile":      "profile",
							"resourceType": "resourceType",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe idps object",
			Input:  []string{"idps"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"idps": {
						DisplayName: "Identity Providers",
						FieldsMap: map[string]string{
							"id":       "id",
							"type":     "type",
							"name":     "name",
							"status":   "status",
							"protocol": "protocol",
							"policy":   "policy",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe authorizationServers object",
			Input:  []string{"authorizationServers"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"authorizationServers": {
						DisplayName: "Authorization Servers",
						FieldsMap: map[string]string{
							"id":          "id",
							"name":        "name",
							"description": "description",
							"audiences":   "audiences",
							"issuer":      "issuer",
							"status":      "status",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe policies object",
			Input:  []string{"policies"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"policies": {
						DisplayName: "Policies",
						FieldsMap: map[string]string{
							"id":          "id",
							"name":        "name",
							"type":        "type",
							"status":      "status",
							"description": "description",
							"priority":    "priority",
							"system":      "system",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe authenticators object",
			Input:  []string{"authenticators"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"authenticators": {
						DisplayName: "Authenticators",
						FieldsMap: map[string]string{
							"id":       "id",
							"key":      "key",
							"name":     "name",
							"type":     "type",
							"status":   "status",
							"settings": "settings",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe multiple objects",
			Input:  []string{"users", "groups", "apps", "logs", "policies"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"id":     "id",
							"status": "status",
						},
					},
					"groups": {
						DisplayName: "Groups",
						FieldsMap: map[string]string{
							"id":   "id",
							"type": "type",
						},
					},
					"apps": {
						DisplayName: "Applications",
						FieldsMap: map[string]string{
							"id":    "id",
							"label": "label",
						},
					},
					"logs": {
						DisplayName: "System Log",
						FieldsMap: map[string]string{
							"uuid":      "uuid",
							"eventType": "eventType",
						},
					},
					"policies": {
						DisplayName: "Policies",
						FieldsMap: map[string]string{
							"id":   "id",
							"name": "name",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: &http.Client{},
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
