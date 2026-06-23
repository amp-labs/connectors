// Package webhook
//
// Temporary noop verifier: while the server-side authenticated client is not
// implemented, this package returns a NoopVerifier from NewVerifier.
// The full Verifier implementation is present and tested, but it is intentionally
// not enabled in production until the server supports authenticated requests.
//
// To enable real verification once server authentication is available:
// - Change NewVerifier to return newVerifier(...) instead of new(NoopVerifier).
package webhook

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// var (
//	ErrFetchSigningKey         = errors.New("failed to fetch signing key to verify webhook event")
//	ErrMissingContentSignature = errors.New("missing x-content-signature header")
//	ErrMissingKeyURL           = errors.New("missing key_url in event body")
//)

type NoopVerifier struct{}

// NewVerifier constructs webhook message verifier.
// ==================================================================================
//
// NOTE: temporary noop verifier!!!
// The server does not currently provide an authenticated HTTP client required by Verifier.
// NewVerifier therefore returns NoopVerifier.
//
// ==================================================================================.
func NewVerifier(client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo, clientID string) *NoopVerifier {
	// At the moment the noop verifier is returned, this must be replaced with newVerifier.
	return new(NoopVerifier)
}

func (v NoopVerifier) VerifyWebhookMessage(
	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	return true, nil
}

// type Verifier struct {
//	client       *common.JSONHTTPClient
//	providerInfo *providers.ProviderInfo
//
//	clientID string
//}
//
// func newVerifier(client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo, clientID string) *Verifier {
//	return &Verifier{
//		client:       client,
//		providerInfo: providerInfo,
//		clientID:     clientID,
//	}
//}
//
//// VerifyWebhookMessage verifies a ConnectWise webhook callback using the
//// x-content-signature header and the signing key referenced by the event's
//// Metadata.key_url field.
////
//// It parses the raw request body to extract key_url, fetches the signing key,
//// computes the expected signature from the raw body bytes, and compares the
//// computed value with the received signature.
////
//// It returns true when the webhook signature is valid. If verification cannot
//// be performed, it returns a non-nil error. If the signature is invalid, it
//// returns (false, nil).
// func (v Verifier) VerifyWebhookMessage(
//	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
// ) (bool, error) {
//	receivedSignature := getHeaderValue(request.Headers, "X-Content-Signature")
//	if receivedSignature == "" {
//		return false, ErrMissingContentSignature
//	}
//
//	var message messageFormat
//	if err := json.Unmarshal(request.Body, &message); err != nil {
//		return false, err
//	}
//
//	keyURL := message.Metadata.KeyURL
//	if keyURL == "" {
//		return false, ErrMissingKeyURL
//	}
//
//	sharedSecretKey, err := v.fetchSigningKey(ctx, keyURL)
//	if err != nil {
//		return false, err
//	}
//
//	shaSum := sha256.Sum256([]byte(sharedSecretKey))
//	mac := hmac.New(sha256.New, shaSum[:])
//	mac.Write(request.Body)
//	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
//
//	if subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(receivedSignature)) != 1 {
//		return false, nil
//	}
//
//	return true, nil
//}
//
// func (v Verifier) fetchSigningKey(ctx context.Context, url string) (string, error) {
//	resp, body, err := v.client.HTTPClient.Get(ctx, url, v.clientIdHeader())
//	if err != nil {
//		return "", fmt.Errorf("%w: %w", ErrFetchSigningKey, err)
//	}
//	defer resp.Body.Close()
//
//	if !httpkit.Status2xx(resp.StatusCode) {
//		return "", fmt.Errorf("%w: HTTP code %v", ErrFetchSigningKey, resp.StatusCode)
//	}
//
//	var response signingKeyResponse
//	if err = json.Unmarshal(body, &response); err != nil {
//		return "", fmt.Errorf("%w: %w", ErrFetchSigningKey, err)
//	}
//
//	return response.SigningKey, nil
//}
//
// type messageFormat struct {
//	Metadata struct {
//		KeyURL string `json:"key_url"`
//	} `json:"Metadata"`
//}
//
//// The event message the webhook will receive will contain "Metadata.key_url".
//// By invoking the GET "<key_url>" the response you get will be signingKeyResponse.
// type signingKeyResponse struct {
//	SigningKey string `json:"signing_key"`
//}
//
// func (v Verifier) clientIdHeader() common.Header {
//	return common.Header{
//		Key:   "ClientId",
//		Value: v.clientID,
//	}
//}
//
// func getHeaderValue(headers http.Header, name string) string {
//	if headers == nil {
//		return ""
//	}
//
//	if value := headers.Get(name); value != "" {
//		return strings.TrimSpace(value)
//	}
//
//	canonicalName := textproto.CanonicalMIMEHeaderKey(name)
//	if values, ok := headers[canonicalName]; ok && len(values) > 0 {
//		return strings.TrimSpace(values[0])
//	}
//
//	if values, ok := headers[strings.ToLower(name)]; ok && len(values) > 0 {
//		return strings.TrimSpace(values[0])
//	}
//
//	return ""
//}
