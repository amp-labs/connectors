package common

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

// InterpretError interprets the given HTTP response (in a fairly straightforward
// way) and returns an error that can be handled by the caller.
func InterpretError(res *http.Response, body []byte) error {
	createError := func(err error) error {
		if len(body) == 0 {
			return err
		} else {
			return fmt.Errorf("%w: %s", err, string(body))
		}
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		// Access token invalid, refresh token and retry
		return NewHTTPStatusError(res.StatusCode, createError(ErrAccessToken))
	case http.StatusForbidden:
		// Forbidden, not retryable
		return NewHTTPStatusError(res.StatusCode, createError(ErrForbidden))
	case http.StatusNotFound:
		// Semantics are debatable (temporarily missing vs. permanently gone), but for now treat this as a retryable error
		return NewHTTPStatusError(res.StatusCode, createError(ErrRetryable))
	case http.StatusTooManyRequests:
		// Too many requests, retryable
		return NewHTTPStatusError(res.StatusCode, createError(ErrRetryable))
	}

	if res.StatusCode >= 400 && res.StatusCode < 500 {
		return NewHTTPStatusError(res.StatusCode, createError(ErrCaller))
	} else if res.StatusCode >= 500 && res.StatusCode < 600 {
		return NewHTTPStatusError(res.StatusCode, createError(ErrServer))
	}

	return NewHTTPStatusError(res.StatusCode, createError(ErrUnknown))
}

type ErrorPostProcessor struct {
	Process func(err error) error
}

// This is the main gateway method that handles errors produced
//   - by http clients JSON, XML, etc.;
//   - by underlying oauth2 library;
//
// Otherwise if the connector has configured a callback it will be called to get
// better error explanation.
func (p ErrorPostProcessor) handleError(err error) error {
	if err == nil {
		return nil
	}

	// Errors may originate from different sources and libraries.
	// Here, we monitor specific errors of interest and enhance them with synonymous errors
	// that can be used in conditional logic by the end caller.
	err = transformOauth2LibraryError(err)

	switch {
	case errors.Is(err, ErrAccessToken):
		slog.Warn("Access token invalid, retrying", "error", err)

		fallthrough
	case errors.Is(err, ErrRetryable):
		fallthrough
	case errors.Is(err, ErrApiDisabled):
		fallthrough
	case errors.Is(err, ErrForbidden):
		fallthrough
	default:
		if p.Process != nil {
			return p.Process(err)
		}

		return err
	}
}

// By default, errors coming from Oauth2 library are mapped here.
// Any future errors that need to be converted to the in-house errors should be added here.
// One important and widespread error is InvalidGrant which happens on invalid refresh token.
func transformOauth2LibraryError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		var oauthErr *oauth2.RetrieveError
		if urlErr != nil && errors.As(urlErr.Err, &oauthErr) {
			if oauthErr.ErrorCode == "invalid_grant" {
				return errors.Join(ErrInvalidGrant, err)
			}
		}
	}

	// otherwise, output is the same as input
	return err
}
