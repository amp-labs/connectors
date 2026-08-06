package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

const (
	stripeSignatureHeader = "Stripe-Signature"
	keyValuePairParts     = 2
	defaultTolerance      = 5 * time.Minute
)

// VerifyWebhookMessage verifies the signature of a webhook message from Stripe.
// Stripe uses HMAC-SHA256 with the format: signed_payload = timestamp + "." + requestBody.
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	if request == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
	}

	verificationParams, err := common.AssertType[*VerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingParams, err)
	}

	if verificationParams.Secret == "" {
		return false, fmt.Errorf("%w: secret cannot be empty", errMissingParams)
	}

	// Determine tolerance: use provided value or default to 5 minutes
	tolerance := verificationParams.Tolerance
	if tolerance == 0 {
		tolerance = defaultTolerance
	} else if tolerance < 0 {
		return false, fmt.Errorf("%w", errInvalidTolerance)
	}

	signatureHeader := request.Headers.Get(stripeSignatureHeader)
	if signatureHeader == "" {
		return false, fmt.Errorf("%w: missing %s header", errMissingSignature, stripeSignatureHeader)
	}

	timestampStr, signatures, err := parseStripeSignature(signatureHeader)
	if err != nil {
		return false, fmt.Errorf("failed to parse Stripe signature: %w", err)
	}

	if err := validateTimestampRecency(timestampStr, tolerance); err != nil {
		return false, err
	}

	return verifySignature(verificationParams.Secret, timestampStr, request.Body, signatures)
}

// verifySignature computes the expected HMAC-SHA256 signature and compares it with the received signatures.
func verifySignature(secret, timestampStr string, body []byte, receivedSignatures []string) (bool, error) {
	// Create signed_payload: timestamp + "." + requestBody
	signedPayload := fmt.Sprintf("%s.%s", timestampStr, string(body))

	// Compute expected signature using HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures using constant-time comparison
	// Check if any of the received signatures matches the expected signature
	for _, sig := range receivedSignatures {
		if hmac.Equal([]byte(expectedSignature), []byte(sig)) {
			return true, nil
		}
	}

	return false, fmt.Errorf("%w: signature mismatch", errInvalidSignature)
}

// validateTimestampRecency validates that the webhook timestamp is within the allowed tolerance.
func validateTimestampRecency(timestampStr string, tolerance time.Duration) error {
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	timestampTime := time.Unix(timestamp, 0)
	now := time.Now()
	timeDiff := now.Sub(timestampTime)

	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	if timeDiff > tolerance {
		if timestampTime.Before(now) {
			return errTimestampTooOld
		}

		return errTimestampTooFarInFuture
	}

	return nil
}

func parseStripeSignature(header string) (string, []string, error) {
	elements := strings.Split(header, ",")

	var timestamp string

	var signatures []string

	for _, element := range elements {
		element = strings.TrimSpace(element)

		parts := strings.SplitN(element, "=", keyValuePairParts)
		if len(parts) != keyValuePairParts {
			continue
		}

		prefix := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch prefix {
		case "t":
			timestamp = value
		case "v1", "v2", "v0":
			signatures = append(signatures, value)
		}
	}

	if timestamp == "" {
		return "", nil, fmt.Errorf("%w", errMissingTimestamp)
	}

	if len(signatures) == 0 {
		return "", nil, fmt.Errorf("%w", errNoSignaturesFound)
	}

	return timestamp, signatures, nil
}
