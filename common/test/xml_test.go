package test

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"
)

func TestXMLData(testing *testing.T) {
	testing.Parallel()

	// Test XMLData.String() with SelfClosing = false
	xmlData := &common.XMLData{
		XMLName:     "test",
		Attributes:  []*common.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []common.XMLSchema{common.XMLString("test")},
		SelfClosing: false,
	}

	xmlStr := xmlData.String()
	assert.Equal(testing, `<test test="test">test</test>`, xmlStr)

	// Test XMLData.String() with SelfClosing = false
	xmlData = &common.XMLData{
		XMLName:     "test",
		Attributes:  []*common.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []common.XMLSchema{common.XMLString("test")},
		SelfClosing: true,
	}

	xmlStr = xmlData.String()
	assert.Equal(testing, `<test test="test"/>`, xmlStr)

	// Test XMLData.String() with no attributes
	xmlData = &common.XMLData{
		XMLName:     "test",
		Attributes:  []*common.XMLAttributes{},
		Children:    []common.XMLSchema{common.XMLString("test")},
		SelfClosing: false,
	}

	xmlStr = xmlData.String()
	assert.Equal(testing, `<test>test</test>`, xmlStr)

	// Test XMLData.String() with no children
	xmlData = &common.XMLData{
		XMLName:     "test",
		Attributes:  []*common.XMLAttributes{{Key: "test", Value: "test"}},
		Children:    []common.XMLSchema{},
		SelfClosing: true,
	}

	xmlStr = xmlData.String()
	assert.Equal(testing, `<test test="test"/>`, xmlStr)
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

	xmlStr := xmlData.String()
	//nolint:lll

	expectation := `<metadata xsi:type="CustomField"><fullName>TestObject13__c.Comments__c</fullName><label>Comments</label></metadata>`

	assert.Equal(testing, expectation, xmlStr)
}

func TestXMLValidation(t *testing.T) {
	t.Parallel()

	type testData struct {
		rawJson             []byte
		data                *common.XMLData
		message             string
		expectUnmarshalFail bool
		valid               bool
	}

	tests := []testData{
		{
			rawJson: []byte(`{
				"xmlName": "fullName",
				"attributes": [{"key": "testkey", "value": "testvalue"}],
				"children": [
					{"xmlName": "fullName"},"TestObject"
				],
				"selfClosing": false
			}`),
			data: &common.XMLData{
				XMLName:     "fullName",
				Attributes:  []*common.XMLAttributes{{Key: "testkey", Value: "testvalue"}},
				Children:    []common.XMLSchema{common.XMLString("TestObject")},
				SelfClosing: false,
			},
			message: "test fully populated XMLData",
			valid:   true,
		},
		{
			rawJson: []byte(`{
			}`),
			data: &common.XMLData{
				XMLName:     "",
				Attributes:  nil,
				Children:    nil,
				SelfClosing: false,
			},
			message: "test empty json, should fail validation",
			valid:   false,
		},
		{
			rawJson: []byte(`{
			}`),
			data: &common.XMLData{
				XMLName:     "",
				Attributes:  nil,
				Children:    nil,
				SelfClosing: false,
			},
			message: "test invalid child valiues, should fail validation",
			valid:   false,
		},
		{
			rawJson: []byte(`{
				"xmlName": "fullName",
				"attributes": null,
				"children": [
				],
				"selfClosing": true
			}`),
			data: &common.XMLData{
				XMLName:     "fullName",
				Attributes:  nil,
				Children:    []common.XMLSchema{},
				SelfClosing: true,
			},
			message: "selfclosing false with no children, should fail validation",
			valid:   true,
		},
		{
			rawJson: []byte(`{
				"xmlName": "fullName<",
				"attributes": null,
				"children": [
				],
				"selfClosing": true
			}`),
			data: &common.XMLData{
				XMLName:     "fullName<",
				Attributes:  nil,
				Children:    []common.XMLSchema{},
				SelfClosing: true,
			},
			message: "invalid xmlName < should fail validation",
			valid:   false,
		},
	}

	for _, test := range tests {
		failName := &common.XMLData{}
		err := json.Unmarshal(test.rawJson, failName)

		if test.expectUnmarshalFail {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)

		err = failName.Validate()

		if test.valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
