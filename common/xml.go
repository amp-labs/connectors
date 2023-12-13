package common

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	openParen           = "<"
	closeParen          = ">"
	closeParenWithSlash = "/>"
)

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
	XMLName     string           `json:"xmlName"`
	Attributes  []*XMLAttributes `json:"attributes"`
	Children    []XMLSchema      `json:"children"`
	SelfClosing bool             `json:"selfClosing"`
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

	if err := json.Unmarshal(data["selfClosing"], &x.SelfClosing); err != nil {
		return err
	}

	return nil
}

func (x *XMLData) ToXML() string {
	start := x.startTag()
	if x.SelfClosing {
		return start
	}

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
