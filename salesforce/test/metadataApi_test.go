package test

import (
	"testing"

	"github.com/amp-labs/connectors/salesforce"
	"github.com/stretchr/testify/require"
)

func TestXMLData(testing *testing.T) {
	testing.Parallel()

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

func TestXMLJSON(testing *testing.T) {
	testing.Parallel()

	rawJSON := []byte(`{
		"xmlName": "metadata",
		"attributes": [
			{
				"key": "xsi:type",
				"value": "CustomField"
			}
		],
		"children": [
			{
				"xmlName": "fullName",
				"attributes": null,
				"children": [
					"TestObject13__c.Comments__c"
				],
				"selfClosing": false
			},
			{
				"xmlName": "label",
				"attributes": null,
				"children": [
					"Comments"
				],
				"selfClosing": false
			}
		],
		"selfClosing": false
	}`)

	xmlData := &salesforce.XMLData{}
	err := xmlData.UnmarshalJSON(rawJSON)

	require.NoError(testing, err)

	xmlStr := xmlData.ToXML()
	//nolint:lll
	if xmlStr != `<metadata xsi:type="CustomField"><fullName>TestObject13__c.Comments__c</fullName><label>Comments</label></metadata>` {
		//nolint:lll
		testing.Errorf("XMLData.ToXML() = %s; want <metadata xsi:type=\"CustomField\"><fullName>TestObject13__c.Comments__c</fullName><label>Comments</label></metadata>", xmlStr)
	}
}
