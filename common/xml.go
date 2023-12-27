package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator"
)

const (
	openParen           = "<"
	closeParen          = ">"
	closeParenWithSlash = "/>"
)

var (
	//nolint:gochecknoglobals
	validate          = validator.New()
	ErrNotXMLChildren = errors.New("children must be of type 'XMLData' or 'XMLString'")
	ErrNoSelfClosing  = errors.New("selfClosing cannot be true if children are not present")
	ErrNoParens       = errors.New("value cannot contain < or >")
)

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

	chilren := []string{}
	for _, child := range x.Children {
		chilren = append(chilren, child.String())
	}

	return fmt.Sprintf("%s%s%s", start, strings.Join(chilren, ""), end)
}

func (x *XMLData) startTag() string {
	attributes := make([]string, len(x.Attributes))
	for i, attr := range x.Attributes {
		attributes[i] = attr.String()
	}

	attrStr := strings.Join(attributes, " ")

	var close string //nolint:predeclared

	if x.SelfClosing {
		close = closeParenWithSlash
	} else {
		close = closeParen
	}

	if attrStr == "" {
		return fmt.Sprintf("%s%s%s", openParen, x.XMLName, close)
	}

	return fmt.Sprintf("%s%s %s%s", openParen, x.XMLName, attrStr, close)
}

func (x *XMLData) endTag() string {
	return fmt.Sprintf("</%s>", x.XMLName)
}
