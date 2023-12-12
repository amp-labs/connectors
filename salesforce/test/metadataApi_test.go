package test

import (
	"testing"

	"github.com/amp-labs/connectors/salesforce"
)

func TestXMLData(testing *testing.T) {

	// Test XMLData.ToXML() with SelfClosing = false
	xmlData := &salesforce.XMLData{
		XMLName:     "test",
		Attributes:  []*salesforce.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []salesforce.XMLSchema{salesforce.XMLString("test")},
		SelfClosing: false,
	}

	xmlStr := xmlData.ToXML()
	if xmlStr != `<test test="test">test</test>` {
		testing.Errorf("XMLData.ToXML() = %s; want <test test=\"test\">test</test>", xmlStr)
	}

	// Test XMLData.ToXML() with SelfClosing = false
	xmlData = &salesforce.XMLData{
		XMLName:     "test",
		Attributes:  []*salesforce.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []salesforce.XMLSchema{salesforce.XMLString("test")},
		SelfClosing: true,
	}

	xmlStr = xmlData.ToXML()
	if xmlStr != `<test test="test"/>` {
		testing.Errorf("XMLData.ToXML() = %s; want <test test=\"test\"/>", xmlStr)
	}

	// Test XMLData.ToXML() with no attributes

	xmlData = &salesforce.XMLData{
		XMLName:     "test",
		Attributes:  []*salesforce.XMLAttributes{},
		Children:    []salesforce.XMLSchema{salesforce.XMLString("test")},
		SelfClosing: false,
	}

	xmlStr = xmlData.ToXML()
	if xmlStr != `<test>test</test>` {
		testing.Errorf("XMLData.ToXML() = %s; want <test>test</test>", xmlStr)
	}

	// Test XMLData.ToXML() with no children

	xmlData = &salesforce.XMLData{
		XMLName:     "test",
		Attributes:  []*salesforce.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []salesforce.XMLSchema{},
		SelfClosing: true,
	}

	xmlStr = xmlData.ToXML()
	if xmlStr != `<test test="test"/>` {
		testing.Errorf("XMLData.ToXML() = %s; want <test test=\"test\"/>", xmlStr)
	}

}
