package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
	"github.com/subchen/go-xmldom"
)

const (
	openParen           = "<"
	closeParen          = ">"
	closeParenWithSlash = "/>"
)

var (
	//nolint:gochecknoglobals
	validate          = validator.New()
	ErrNoXMLRoot      = errors.New("xml document has no root")
	ErrNotXMLChildren = errors.New("children must be of type 'XMLData' or 'XMLString'")
	ErrNoSelfClosing  = errors.New("selfClosing cannot be true if children are not present")
	ErrNoParens       = errors.New("value cannot contain < or >")
)

// XMLHTTPClient speaks from http client in XML.
type XMLHTTPClient struct {
	HTTPClient *HTTPClient // underlying HTTP client. Required.
}

type XMLHTTPResponse struct {
	// bodyBytes is the raw response body.
	bodyBytes []byte

	// Code is the HTTP status code of the response.
	Code int

	// Headers are the HTTP headers of the response.
	Headers http.Header

	// Body is the unmarshalled response body in XML form. Content is the same as bodyBytes
	Body *xmldom.Document
}

func (r XMLHTTPResponse) GetRoot() (*xmldom.Node, error) {
	if r.Body == nil || r.Body.Root == nil {
		return nil, ErrNoXMLRoot
	}

	return r.Body.Root, nil
}

// Get makes a GET request to the given URL and returns the response body as a XML object.
// If the response is not a 2xx, an error is returned. If the response is a 401, the caller should
// refresh the access token and retry the request. If errorHandler is nil, then the default error
// handler is used. If not, the caller can inject their own error handling logic.
func (j *XMLHTTPClient) Get(ctx context.Context, url string, headers ...Header) (*XMLHTTPResponse, error) {
	res, body, err := j.HTTPClient.Get(ctx, url, addAcceptXMLHeader(headers)) //nolint:bodyclose
	if err != nil {
		return nil, err
	}

	return parseXMLResponse(res, body)
}

// parseXMLResponse parses the given HTTP response and returns a XMLHTTPResponse.
func parseXMLResponse(res *http.Response, body []byte) (*XMLHTTPResponse, error) {
	if len(body) == 0 {
		// Empty XML response is not allowed
		return nil, ErrNotXML
	}
	// Ensure the response is XML
	ct := res.Header.Get("Content-Type")
	if len(ct) > 0 {
		mimeType, _, err := mime.ParseMediaType(ct)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content type: %w", err)
		}

		if mimeType != "application/xml" {
			return nil, fmt.Errorf("%w: expected content type to be application/xml, got %s", ErrNotXML, mimeType)
		}
	}

	// Unmarshall the response body into XML
	xmlBody, err := xmldom.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, NewHTTPStatusError(res.StatusCode, fmt.Errorf("failed to unmarshall response body into XML: %w", err))
	}

	return &XMLHTTPResponse{
		bodyBytes: body,
		Code:      res.StatusCode,
		Headers:   res.Header,
		Body:      xmlBody,
	}, nil
}

func addAcceptXMLHeader(headers []Header) []Header {
	if headers == nil {
		headers = make([]Header, 0)
	}

	return append(headers, Header{Key: "Accept", Value: "application/xml"})
}

type XMLSchema interface {
	String() string
	Validate() error
}

type XMLAttributes struct {
	Key   string `json:"key"   validate:"required,excludesall=<>,excludesrune=<>"`
	Value string `json:"value" validate:"excludesall=<>,excludesrune=<>"`
}

func (attr *XMLAttributes) String() string {
	return fmt.Sprintf(`%s="%s"`, attr.Key, attr.Value)
}

func (attr *XMLAttributes) Validate() error {
	if err := validate.Struct(attr); err != nil {
		return err
	}

	return nil
}

type XMLString string

func (str XMLString) Validate() error {
	if strings.Contains(string(str), "<") || strings.Contains(string(str), ">") {
		return fmt.Errorf("XMLString %w", ErrNoParens)
	}

	return nil
}

func (str XMLString) String() string {
	return string(str)
}

type XMLData struct {
	XMLName     string           `json:"xmlName,omitempty"     validate:"required,excludesall=<>"`
	Attributes  []*XMLAttributes `json:"attributes,omitempty"`
	Children    []XMLSchema      `json:"children,omitempty"`
	SelfClosing bool             `json:"selfClosing,omitempty"`
}

func (x *XMLData) Validate() error {
	if err := validate.Struct(x); err != nil {
		return err
	}

	if x.SelfClosing && len(x.Children) > 0 {
		return ErrNoSelfClosing
	}

	if x.Children != nil {
		for _, child := range x.Children {
			if err := child.Validate(); err != nil {
				return err
			}
		}
	}

	if x.Attributes != nil {
		for _, attr := range x.Attributes {
			if err := validate.Struct(attr); err != nil {
				return err
			}
		}
	}

	return nil
}

//nolint:cyclop
func (x *XMLData) UnmarshalJSON(b []byte) error {
	data := make(map[string]*json.RawMessage)
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	if data["xmlName"] != nil {
		if err := json.Unmarshal(*data["xmlName"], &x.XMLName); err != nil {
			return err
		}
	}

	if data["selfClosing"] != nil {
		if err := json.Unmarshal(*data["selfClosing"], &x.SelfClosing); err != nil {
			return err
		}
	}

	if data["attributes"] != nil {
		attributes := []*XMLAttributes{}
		if err := json.Unmarshal(*data["attributes"], &attributes); err != nil {
			return err
		}

		x.Attributes = attributes
	}

	//nolin:nestif
	if data["children"] != nil {
		children := []*json.RawMessage{}

		if err := json.Unmarshal(*data["children"], &children); err != nil {
			return err
		}

		x.Children = make([]XMLSchema, len(children))

		for idx, child := range children {
			var childData *XMLData

			errXML := json.Unmarshal(*child, &childData)
			if errXML == nil {
				x.Children[idx] = childData

				continue
			}

			var xmlString XMLString

			errString := json.Unmarshal(*child, &xmlString)
			if errString == nil {
				x.Children[idx] = xmlString

				continue
			}

			return fmt.Errorf("%w: %s", ErrNotXMLChildren, string(*child))
		}
	}

	return nil
}

func (x *XMLData) String() string {
	start := x.startTag()
	if x.SelfClosing {
		return start
	}

	end := x.endTag()

	children := make([]string, 0)
	for _, child := range x.Children {
		children = append(children, child.String())
	}

	return fmt.Sprintf("%s%s%s", start, strings.Join(children, ""), end)
}

func (x *XMLData) startTag() string {
	attributes := make([]string, len(x.Attributes))
	for i, attr := range x.Attributes {
		attributes[i] = attr.String()
	}

	attrStr := strings.Join(attributes, " ")

	var closingTag string //nolint:predeclared

	if x.SelfClosing {
		closingTag = closeParenWithSlash
	} else {
		closingTag = closeParen
	}

	if attrStr == "" {
		return fmt.Sprintf("%s%s%s", openParen, x.XMLName, closingTag)
	}

	return fmt.Sprintf("%s%s %s%s", openParen, x.XMLName, attrStr, closingTag)
}

func (x *XMLData) endTag() string {
	return fmt.Sprintf("</%s>", x.XMLName)
}
