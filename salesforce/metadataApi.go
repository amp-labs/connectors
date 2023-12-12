package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

const (
	versionPrefix             = "v"
	openParenthesis           = "<"
	closeParenthesis          = ">"
	closeWithSlashParenthesis = "/>"
)

var ErrCreateMetadata = fmt.Errorf("error in CreateMetadata")

type XMLSchema interface {
	ToXML() string
}

type XMLAttributes struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (x *XMLAttributes) ToXML() string {
	return fmt.Sprintf(`%s="%s"`, x.Key, x.Value)
}

type XMLString string

func (x XMLString) ToXML() string {
	return string(x)
}

type XMLData struct {
	XMLName      string           `json:"xmlName"`
	Attributes   []*XMLAttributes `json:"attributes"`
	Children     []XMLSchema      `json:"children"`
	HasEndingTag bool             `json:"hasEndingTag"`
}

//nolint:cyclop
func (x *XMLData) UnmarshalJSON(b []byte) error {
	data := make(map[string]json.RawMessage)
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	if err := json.Unmarshal(data["xmlName"], &x.XMLName); err != nil {
		return err
	}

	if err := json.Unmarshal(data["attributes"], &x.Attributes); err != nil {
		return err
	}

	children := []interface{}{}
	if err := json.Unmarshal(data["children"], &children); err != nil {
		return err
	}

	for _, child := range children {
		if childValue, ok := child.(string); ok {
			x.Children = append(x.Children, XMLString(childValue))

			continue
		}

		if childValue, ok := child.(map[string]interface{}); ok {
			childData, err := json.Marshal(childValue)
			if err != nil {
				return err
			}

			childXML := &XMLData{}
			if err := json.Unmarshal(childData, childXML); err != nil {
				return err
			}

			x.Children = append(x.Children, childXML)

			continue
		}
	}

	if err := json.Unmarshal(data["hasEndingTag"], &x.HasEndingTag); err != nil {
		return err
	}

	return nil
}

func (x *XMLData) ToXML() string {
	start := x.startTag()
	end := x.endTag()

	chilren := []string{}
	for _, child := range x.Children {
		chilren = append(chilren, child.ToXML())
	}

	return fmt.Sprintf("%s%s%s", start, strings.Join(chilren, ""), end)
}

func (x *XMLData) startTag() string {
	attributes := make([]string, len(x.Attributes))
	for i, attr := range x.Attributes {
		attributes[i] = attr.ToXML()
	}

	attrStr := strings.Join(attributes, " ")

	var close string //nolint:predeclared

	if !x.HasEndingTag {
		close = closeWithSlashParenthesis
	} else {
		close = closeParenthesis
	}

	if attrStr == "" {
		return fmt.Sprintf("%s%s%s", openParenthesis, x.XMLName, close)
	}

	return fmt.Sprintf("%s%s %s%s", openParenthesis, x.XMLName, attrStr, close)
}

func (x *XMLData) endTag() string {
	if !x.HasEndingTag {
		return ""
	}

	return fmt.Sprintf("</%s>", x.XMLName)
}

func (c *Connector) CreateMetadata(ctx context.Context, operation *XMLData, accessToken string) (string, error) {
	data := preparePayload(operation, accessToken)

	endPointURL, err := url.JoinPath(c.Client.Base, "services/Soap/m/"+removePrefix(c.APIVersion(), versionPrefix))
	if err != nil {
		return "", err
	}

	client := c.Client.Client

	byteData := []byte(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endPointURL, bytes.NewBuffer(byteData))
	if err != nil {
		return "", err
	}

	req.ContentLength = int64(len(byteData))

	addSOAPAPIHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	defer func() {
		if res != nil && res.Body != nil {
			if closeErr := res.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	fmt.Println(string(body))
	fmt.Println(res.StatusCode)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("%w: %s", ErrCreateMetadata, string(body))
	}

	return string(body), nil
}

func removePrefix(s string, prefix string) string {
	if !strings.HasPrefix(s, prefix) {
		return s
	}

	return s[len(prefix):]
}

func addSOAPAPIHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "text/xml")
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

func formOperationXML(oper *XMLData) string {
	return oper.ToXML()
}

func preparePayload(oper *XMLData, accessToken string) string {
	metadata := formOperationXML(oper)
	header := getHeader([]string{getSessionHeader(accessToken)})
	body := getBody([]string{metadata})
	data := getEnvelope(header, body)

	return data
}

func GetTokenUpdater(tok *oauth2.Token) common.OAuthOption {
	// Whenever a token is updated, we want to persist the new access+refresh token
	return common.WithTokenUpdated(func(oldToken, newToken *oauth2.Token) error {
		tok.AccessToken = newToken.AccessToken
		tok.RefreshToken = newToken.RefreshToken
		tok.TokenType = newToken.TokenType
		tok.Expiry = newToken.Expiry

		return nil
	})
}
