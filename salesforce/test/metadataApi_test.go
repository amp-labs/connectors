package test

import (
	"testing"

	"github.com/amp-labs/connectors/common"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

func TestXMLData(testing *testing.T) {
	testing.Parallel()

	// Test XMLData.ToXML() with SelfClosing = false
	xmlData := &common.XMLData{
		XMLName:     "test",
		Attributes:  []*common.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []common.XMLSchema{common.XMLString("test")},
		SelfClosing: false,
	}

	xmlStr := xmlData.ToXML()
	if xmlStr != `<test test="test">test</test>` {
		testing.Errorf("XMLData.ToXML() = %s; want <test test=\"test\">test</test>", xmlStr)
	}

	// Test XMLData.ToXML() with SelfClosing = false
	xmlData = &common.XMLData{
		XMLName:     "test",
		Attributes:  []*common.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []common.XMLSchema{common.XMLString("test")},
		SelfClosing: true,
	}

	xmlStr = xmlData.ToXML()
	if xmlStr != `<test test="test"/>` {
		testing.Errorf("XMLData.ToXML() = %s; want <test test=\"test\"/>", xmlStr)
	}

	// Test XMLData.ToXML() with no attributes

	xmlData = &common.XMLData{
		XMLName:     "test",
		Attributes:  []*common.XMLAttributes{},
		Children:    []common.XMLSchema{common.XMLString("test")},
		SelfClosing: false,
	}

	xmlStr = xmlData.ToXML()
	if xmlStr != `<test>test</test>` {
		testing.Errorf("XMLData.ToXML() = %s; want <test>test</test>", xmlStr)
	}

	// Test XMLData.ToXML() with no children

	xmlData = &common.XMLData{
		XMLName:     "test",
		Attributes:  []*common.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []common.XMLSchema{},
		SelfClosing: true,
	}

	xmlStr = xmlData.ToXML()

	expectation := `<test test="test"/>`

	assert.Equal(testing, expectation, xmlStr)
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

	xmlData := &common.XMLData{}
	err := xmlData.UnmarshalJSON(rawJSON)

	require.NoError(testing, err)

	xmlStr := xmlData.ToXML()
	//nolint:lll

	expectation := `<metadata xsi:type="CustomField"><fullName>TestObject13__c.Comments__c</fullName><label>Comments</label></metadata>`

	assert.Equal(testing, expectation, xmlStr)
}
