package webhook

//func TestVerifyWebhookMessage(t *testing.T) {
//	t.Parallel()
//
//	const signingKey = "test-signing-key"
//	eventMessage := testutils.DataFromFile(t, "contact-create.json")
//
//	tests := []testroutines.TestCaseWebhookMessageVerification{
//		{
//			Name: "Valid signature",
//			Input: testroutines.WebhookMessageVerificationParams{
//				Request: &common.WebhookRequest{Body: eventMessage},
//			},
//			InputMutator: withSignatureHeader(signingKey),
//			Server: mockserver.Conditional{
//				Setup: mockserver.ContentText(),
//				If: mockcond.And{
//					mockcond.Path("/v2025_1//system/callbackkey/index.rails"),
//					mockcond.QueryParam("companyName", "myCompany"),
//					mockcond.QueryParam("id", "89966c04-9889-46ad-8d5e-421103aca922"),
//					mockcond.Header(http.Header{"ClientId": []string{"test-client-id"}}),
//				},
//				Then: mockserver.ResponseChainedFuncs(
//					mockserver.ContentText(),
//					mockserver.ResponseString(http.StatusOK, `{"signing_key": "test-signing-key"}`),
//				),
//			}.Server(),
//			Expected: true,
//		},
//		{
//			Name: "Invalid signature",
//			Input: testroutines.WebhookMessageVerificationParams{
//				Request: &common.WebhookRequest{
//					Headers: http.Header{"x-content-signature": []string{"mismatching-signature-from-provider"}},
//					Body:    eventMessage,
//				},
//			},
//			InputMutator: replaceServerURLInBody,
//			Server: mockserver.Conditional{
//				Setup: mockserver.ContentText(),
//				If: mockcond.And{
//					mockcond.Path("/v2025_1//system/callbackkey/index.rails"),
//					mockcond.QueryParam("companyName", "myCompany"),
//					mockcond.QueryParam("id", "89966c04-9889-46ad-8d5e-421103aca922"),
//					mockcond.Header(http.Header{"ClientId": []string{"test-client-id"}}),
//				},
//				Then: mockserver.ResponseChainedFuncs(
//					mockserver.ContentText(),
//					mockserver.ResponseString(http.StatusOK, `{"signing_key": "test-signing-key"}`),
//				),
//			}.Server(),
//			Expected: false,
//		},
//		{
//			Name: "Missing signature header in input",
//			Input: testroutines.WebhookMessageVerificationParams{
//				Request: &common.WebhookRequest{Body: eventMessage},
//			},
//			InputMutator: replaceServerURLInBody,
//			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//				_ = json.NewEncoder(w).Encode(signingKeyResponse{
//					SigningKey: signingKey,
//				})
//			})),
//			Expected:     false,
//			ExpectedErrs: []error{testutils.StringError("missing x-content-signature header")},
//		},
//		{
//			Name: "Fetch signing key fails",
//			Input: testroutines.WebhookMessageVerificationParams{
//				Request: &common.WebhookRequest{Body: eventMessage},
//			},
//			InputMutator: withSignatureHeader(signingKey),
//			Server: mockserver.Fixed{
//				Setup:  mockserver.ContentText(),
//				Always: mockserver.Response(http.StatusInternalServerError),
//			}.Server(),
//			Expected:     false,
//			ExpectedErrs: []error{ErrFetchSigningKey},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.Name, func(t *testing.T) {
//			t.Parallel()
//
//			tt.Run(t, func() (testroutines.TestableWebhookMessageVerifier, error) {
//				return constructTestVerifier(tt.Server)
//			})
//		})
//	}
//}
//
//func constructTestVerifier(server *httptest.Server) (*Verifier, error) {
//	transport, err := components.NewTransport(providers.ConnectWise, common.ConnectorParams{
//		AuthenticatedClient: server.Client(),
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	transport.SetUnitTestMockServerBaseURL(server.URL)
//
//	verifier := NewVerifier(transport.JSONHTTPClient(), transport.ProviderInfo(), "test-client-id")
//
//	return verifier, nil
//}
//
//func replaceServerURLInBody(
//	server *httptest.Server, input testroutines.WebhookMessageVerificationParams,
//) testroutines.WebhookMessageVerificationParams {
//	// Body has the URL which should be stubbed so connector will make a call to the mock server.
//	input.Request.Body = []byte(
//		testroutines.ResolveTestServerURL(string(input.Request.Body), server.URL),
//	)
//	return input
//}
//
//func withSignatureHeader(secretKey string) testroutines.InputMutator[testroutines.WebhookMessageVerificationParams] {
//	return func(
//		server *httptest.Server, input testroutines.WebhookMessageVerificationParams,
//	) testroutines.WebhookMessageVerificationParams {
//		input = replaceServerURLInBody(server, input)
//
//		// The body changes between test runs so we must compute the signature ourselves.
//		shaSum := sha256.Sum256([]byte(secretKey))
//		mac := hmac.New(sha256.New, shaSum[:])
//		mac.Write(input.Request.Body)
//		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
//
//		if input.Request.Headers == nil {
//			input.Request.Headers = make(http.Header)
//		}
//		input.Request.Headers.Set("X-Content-Signature", signature)
//
//		return input
//	}
//}
