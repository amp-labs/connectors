package aws

import (
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

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseTrustedTokenIssuersCreate := testutils.DataFromFile(t, "write/issuer/new.json")
	responseUserCreate := testutils.DataFromFile(t, "write/user/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "TrustedTokenIssuers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "Orders", RecordData: "dummy"},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Create TrustedTokenIssuers",
			Input: common.WriteParams{ObjectName: "TrustedTokenIssuers", RecordData: map[string]string{}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME(core.Mime),
				If: mockcond.And{
					mockcond.Header(http.Header{
						"X-Amz-Target": []string{"SWBExternalService.CreateTrustedTokenIssuer"},
					}),
					mockcond.Body(`{"InstanceArn":"test-instance-arn"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseTrustedTokenIssuersCreate),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "arn:aws:sso::471112904037:trustedTokenIssuer/ssoins-668432c2a02ced8f/tti-014b2500-f0f0-70b7-ac48-ceddaf77c5fa", // nolint:lll
				Errors:   nil,
				Data: map[string]any{
					"TrustedTokenIssuerArn": "arn:aws:sso::471112904037:trustedTokenIssuer/ssoins-668432c2a02ced8f/tti-014b2500-f0f0-70b7-ac48-ceddaf77c5fa", // nolint:lll
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update TrustedTokenIssuers",
			Input: common.WriteParams{
				ObjectName: "TrustedTokenIssuers",
				RecordId:   "123456789",
				RecordData: map[string]string{},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME(core.Mime),
				If: mockcond.And{
					mockcond.Header(http.Header{
						"X-Amz-Target": []string{"SWBExternalService.UpdateTrustedTokenIssuer"},
					}),
					mockcond.Body(`{"InstanceArn":"test-instance-arn","TrustedTokenIssuerArn":"123456789"}`),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{}`),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Errors:   nil,
				Data:     map[string]any{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create User",
			Input: common.WriteParams{ObjectName: "Users", RecordData: map[string]string{}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME(core.Mime),
				If: mockcond.And{
					mockcond.Header(http.Header{
						"X-Amz-Target": []string{"AWSIdentityStore.CreateUser"},
					}),
					mockcond.Body(`{"IdentityStoreId":"test-identity-store-id"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseUserCreate),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "11cba580-d011-7007-0ab2-963157a0ffd7",
				Errors:   nil,
				Data: map[string]any{
					"IdentityStoreId": "d-9a670e6550",
					"UserId":          "11cba580-d011-7007-0ab2-963157a0ffd7",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update User",
			Input: common.WriteParams{
				ObjectName: "Users",
				RecordId:   "123456789",
				RecordData: map[string]string{},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME(core.Mime),
				If: mockcond.And{
					mockcond.Header(http.Header{
						"X-Amz-Target": []string{"AWSIdentityStore.UpdateUser"},
					}),
					mockcond.Body(`{"IdentityStoreId":"test-identity-store-id","UserId":"123456789"}`),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{}`),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Errors:   nil,
				Data:     map[string]any{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
