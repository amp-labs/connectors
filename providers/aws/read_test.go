package aws

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/aws/internal/core"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "read/err-not-supported.json")
	responseGroups := testutils.DataFromFile(t, "read/groups.json")
	responseInstances := testutils.DataFromFile(t, "read/instances.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "Groups"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "Groups", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentMIME(core.Mime),
				Always: mockserver.Response(http.StatusBadRequest, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("The operation is not supported for this Identity Center instance"),
			},
		},
		{
			Name: "Service IdentityStore returns Groups",
			Input: common.ReadParams{
				ObjectName: "Groups",
				Fields:     connectors.Fields("GroupId", "DisplayName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME(core.Mime),
				If: mockcond.And{
					mockcond.Header(http.Header{
						"X-Amz-Target": []string{"AWSIdentityStore.ListGroups"},
					}),
					mockcond.Body(`{"IdentityStoreId":"test-identity-store-id"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseGroups),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"groupid":     "01cbe580-e041-7044-a154-832eab370f7a",
						"displayname": "Engineering",
					},
					Raw: map[string]any{
						"IdentityStoreId": "d-9a670e6550",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Service SSO-Admin returns Instances",
			Input: common.ReadParams{
				ObjectName: "Instances",
				Fields:     connectors.Fields("InstanceArn"),
				NextPage:   "somePreviousToken",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME(core.Mime),
				If: mockcond.And{
					mockcond.Header(http.Header{
						"X-Amz-Target": []string{"SWBExternalService.ListInstances"},
					}),
					mockcond.Body(`{"NextToken":"somePreviousToken","InstanceArn":"test-instance-arn"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseInstances),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"instancearn": "arn:aws:sso:::instance/ssoins-668432c2a02ced8f",
					},
					Raw: map[string]any{
						"IdentityStoreId": "d-9a670e6550",
						"OwnerAccountId":  "471112904037",
					},
				}},
				NextPage: "AAMA-EFRSURBSGg1SWlhM3N2aTFLVjl0K1k4eDhHaGxDWTlZdnNmRldtUFl5b2hXeDVZVnV3RmVBL0wwS2pGRm1BQkVxYnhGQm9VaEFBQUFmakI4QmdrcWhraUc5dzBCQndhZ2J6QnRBZ0VBTUdnR0NTcUdTSWIzRFFFSEFUQWVCZ2xnaGtnQlpRTUVBUzR3RVFRTXArK0JMVERpOVBXcDQwWFVBZ0VRZ0R0dEtDRjBnS0xXaHhBQUkxdGlnOE9sT3c5K1ZDMHFVcS9TTkZPMWJpc1V6MlJFSkFuREhqTjA1aHQ2dGF2YXNvb0JDZi8rOEhSSzZhcEozQT091otgI5gnLV-Huqrp8MTjDuPmttTgsLqqh9S6A29al92lcQ9Wz2Z6PaKJM7VbkFZm78S1HC_gEIbp2Ot8OUCcgEkX_Z6cgMHo0XFfOr84", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
