package salesforce

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/subchen/go-xmldom"
	"golang.org/x/oauth2"
)

var (
	ErrCreateMetadata  = errors.New("error in CreateMetadata")
	ErrCreatingRequest = errors.New("error in creating request")
)

func (c *Connector) CreateMetadata(
	ctx context.Context,
	metadata *xmldom.Node,
	tok *oauth2.Token,
) (string, error) {
	req, err := c.prepareXMLRequest(ctx, metadata, tok)
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
		req, err = c.prepareXMLRequest(ctx, metadata, tok)
		if err != nil {
			return "", errors.Join(ErrCreateMetadata, err)
		}
		//nolint:bodyclose,ineffassign,staticcheck,wastedassign
		res, body, err = c.makeRequest(req)
	}

	if err != nil {
		return "", fmt.Errorf("%w: %s", errors.Join(ErrCreateMetadata, err), string(body))
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("%w: %s", ErrCreateMetadata, string(body))
	}

	return string(body), nil
}

func (c *Connector) prepareXMLRequest(
	ctx context.Context,
	metadata *xmldom.Node,
	tok *oauth2.Token,
) (*http.Request, error) {
	data := preparePayload(metadata, tok.AccessToken)

	url, err := c.getDomainURL("services/Soap/m/" + APIVersionSOAP())
	if err != nil {
		return nil, errors.Join(ErrCreatingRequest, err)
	}

	byteData := []byte(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewBuffer(byteData))
	if err != nil {
		return nil, errors.Join(ErrCreatingRequest, err)
	}

	addSOAPHeaders(req)
	req.ContentLength = int64(len(byteData))

	return req, nil
}

func (c *Connector) makeRequest(req *http.Request) (*http.Response, []byte, error) {
	res, err := c.Client.HTTPClient.Client.Do(req)
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

func addSOAPHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "text/xml")
	// SOAP API definition specifies that SOAPAction header should be empty string
	// but if we set to "", API will error
	// so we use "''" instead as workaround
	req.Header.Set("SOAPAction", "''")
}

func getEnvelope(header *xmldom.Node, body *xmldom.Node) *xmldom.Document {
	envelop := xmldom.NewDocument("soapenv:Envelope")
	envelop.Root.SetAttributeValue("xmlns:soapenv", "http://schemas.xmlsoap.org/soap/envelope/")
	envelop.Root.SetAttributeValue("xmlns:xsd", "http://www.w3.org/2001/XMLSchema")
	envelop.Root.SetAttributeValue("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	envelop.Root.Children = []*xmldom.Node{header, body}

	return envelop
}

func getHeader(headers ...*xmldom.Node) *xmldom.Node {
	header := &xmldom.Node{
		Name:     "soapenv:Header",
		Children: headers,
	}
	header.SetAttributeValue("xmlns", "http://soap.sforce.com/2006/04/metadata")

	return header
}

func getSessionHeader(token string) *xmldom.Node {
	sessionId := &xmldom.Node{
		Name: "sessionId",
		Text: token,
	}
	header := &xmldom.Node{
		Name: "SessionHeader",
		Children: []*xmldom.Node{
			sessionId,
		},
	}

	return header
}

func getBody(items ...*xmldom.Node) *xmldom.Node {
	body := &xmldom.Node{
		Name: "soapenv:Body",
		Attributes: []*xmldom.Attribute{
			{
				Name:  "xmlns",
				Value: "http://soap.sforce.com/2006/04/metadata",
			},
		},
		Children: items,
	}

	return body
}

func preparePayload(metadata *xmldom.Node, accessToken string) string {
	sessionHeader := getSessionHeader(accessToken)
	allOrNonHeader := getAllOrNoneHeader(true)
	header := getHeader(allOrNonHeader, sessionHeader)
	body := getBody(metadata)
	envelop := getEnvelope(header, body)

	return envelop.XML()
}

type BoolString string

const (
	True  string = "true"
	False string = "false"
)

func getAllOrNoneHeader(allOrNon bool) *xmldom.Node {
	allOrNonText := False
	if allOrNon {
		allOrNonText = True
	}

	header := &xmldom.Node{
		Name: "AllOrNoneHeader",
		Children: []*xmldom.Node{
			{
				Name: "allOrNone",
				Text: allOrNonText,
			},
		},
	}

	return header
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
