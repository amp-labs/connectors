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
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "Users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Successful User removal",
			Input: common.DeleteParams{ObjectName: "Users", RecordId: "123"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME(core.Mime),
				If: mockcond.And{
					mockcond.Header(http.Header{
						"X-Amz-Target": []string{"AWSIdentityStore.DeleteUser"},
					}),
					mockcond.Body(`{"IdentityStoreId":"test-identity-store-id","UserId":"123"}`),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful Trusted Token Issuer removal",
			Input: common.DeleteParams{ObjectName: "TrustedTokenIssuers", RecordId: "123"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME(core.Mime),
				If: mockcond.And{
					mockcond.Header(http.Header{
						"X-Amz-Target": []string{"SWBExternalService.DeleteTrustedTokenIssuer"},
					}),
					mockcond.Body(`{"InstanceArn":"test-instance-arn","TrustedTokenIssuerArn":"123"}`),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
