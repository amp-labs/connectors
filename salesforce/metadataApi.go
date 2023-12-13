package salesforce

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

var (
	ErrCreateMetadata  = fmt.Errorf("error in CreateMetadata")
	ErrCreatingRequest = fmt.Errorf("error in creating request")
)

func (c *Connector) CreateMetadata(
	ctx context.Context,
	metaDefinition *common.XMLData,
	tok *oauth2.Token,
) (string, error) {
	req, err := c.prepareXMLRequest(ctx, metaDefinition, tok)
	if err != nil {
		return "", err
	}

	res, body, err := c.makeRequest(req) //nolint:bodyclose
	// below is a workaround to refresh token if it is expired
	// normally oauth2 library should handle this
	// but SOAP API does not take token in header
	// but takes it in body
	// So in case of 500 error and INVALID_SESSION_ID in body
	// we know it is session expired, and automatically refresh the token
	// tok object will be updated with new token automatically after failing first call
	// we simply make another call with updated token.
	if res.StatusCode == 500 && strings.Contains(string(body), "INVALID_SESSION_ID") {
		req, err := c.prepareXMLRequest(ctx, metaDefinition, tok)
		if err != nil {
			return "", errors.Join(ErrCreateMetadata, err)
		}

		res, body, err = c.makeRequest(req)
		if err != nil {
			return string(body), fmt.Errorf("%w: %s", ErrCreateMetadata, string(body))
		}
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("%w: %s", ErrCreateMetadata, string(body))
	}

	return string(body), nil
}

func (c *Connector) prepareXMLRequest(
	ctx context.Context,
	operation *common.XMLData,
	tok *oauth2.Token,
) (*http.Request, error) {
	data := preparePayload(operation, tok.AccessToken)

	endPointURL, err := url.JoinPath(c.Client.Base, "services/Soap/m/"+c.APIVersionSOAP())
	if err != nil {
		return nil, errors.Join(ErrCreatingRequest, err)
	}

	byteData := []byte(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endPointURL, bytes.NewBuffer(byteData))
	if err != nil {
		return nil, errors.Join(ErrCreatingRequest, err)
	}

	addSOAPHeaders(req)
	req.ContentLength = int64(len(byteData))

	return req, nil
}

func (c *Connector) makeRequest(req *http.Request) (*http.Response, []byte, error) {
	res, err := c.Client.Client.Do(req)
	if err != nil {
		return res, nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res, nil, err
	}

	defer func() {
		if res != nil && res.Body != nil {
			if closeErr := res.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	return res, body, nil
}

func removePrefix(s string, prefix string) string {
	if !strings.HasPrefix(s, prefix) {
		return s
	}

	return s[len(prefix):]
}

func addSOAPHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "text/xml")
	// SOAP API definition specifies taht SOAPAction header should be empty string
	// but if we set to "", API will error
	// so we use "''" instead as workaround
	req.Header.Set("SOAPAction", "''")
}

func getEnvelope(header string, body string) string {
	return fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?>
		<soapenv:Envelope
			xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
			xmlns:xsd="http://www.w3.org/2001/XMLSchema"
			xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
			%s
			%s
		</soapenv:Envelope>`,
		header,
		body,
	)
}

func getHeader(headers []string) string {
	return fmt.Sprintf(
		`<soapenv:Header xmlns="http://soap.sforce.com/2006/04/metadata">
			<AllOrNoneHeader>
				<allOrNone>true</allOrNone>
			</AllOrNoneHeader>
			%s
		</soapenv:Header>`, strings.Join(headers, ""))
}

func getSessionHeader(token string) string {
	return fmt.Sprintf(
		`<SessionHeader>
		<sessionId>%s</sessionId>
	</SessionHeader>`, token)
}

func getBody(items []string) string {
	return fmt.Sprintf(
		`<soapenv:Body xmlns="http://soap.sforce.com/2006/04/metadata">
			%s
		</soapenv:Body>`, strings.Join(items, ""))
}

func formOperationXML(oper *common.XMLData) string {
	return oper.ToXML()
}

func preparePayload(oper *common.XMLData, accessToken string) string {
	metadata := formOperationXML(oper)
	header := getHeader([]string{getSessionHeader(accessToken)})
	body := getBody([]string{metadata})
	data := getEnvelope(header, body)

	return data
}

func GetTokenUpdater(tok *oauth2.Token) common.OAuthOption {
	// Whenever a token is updated, we want to persist the new access+refresh token
	return common.WithTokenUpdated(func(oldToken, newToken *oauth2.Token) error {
		// this triggeres first API call to metadata API
		// since metadata API doesn't take access token in header
		// we need to update the token manually
		// then make the call again
		tok.AccessToken = newToken.AccessToken
		tok.RefreshToken = newToken.RefreshToken
		tok.TokenType = newToken.TokenType
		tok.Expiry = newToken.Expiry

		return nil
	})
}
