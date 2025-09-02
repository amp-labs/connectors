package custom

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Parses XML output from Salesforce.
// It may include errors for each field.
// See sample response below:
/*
   <upsertMetadataResponse>
       <result>
           <created>false</created>
           <fullName>TestObject15__c.Birthday__c</fullName>
           <success>true</success>
       </result>
       <result>
           <created>true</created>
           <fullName>TestObject15__c.Hobby__c</fullName>
           <success>true</success>
       </result>
   </upsertMetadataResponse>
*/
func parseResponse(response []byte) (*common.UpsertMetadataResult, error) {
	var envelope Envelope

	if err := xml.Unmarshal(response, &envelope); err != nil {
		return nil, err
	}

	errorMessages := datautils.NewStringSet()
	fields := make(map[string]map[string]common.FieldUpsertResult)

	for _, result := range envelope.Body.Response.Results {
		for _, errorObj := range result.Errors {
			errorMessages.AddOne(errorObj.Message)
		}

		parts := strings.Split(result.FullName, ".")
		if len(parts) != 2 { // nolint:mnd
			// Format of the full name must be `ObjectName.FieldName`.
			// Omit this record.
			continue
		}

		objectName, fieldName := parts[0], parts[1]

		action := common.UpsertMetadataActionUpdate
		if result.Created {
			action = common.UpsertMetadataActionCreate
		}

		fieldsMap, ok := fields[objectName]
		if !ok {
			fields[objectName] = make(map[string]common.FieldUpsertResult)
			fieldsMap = fields[objectName]
		}

		fieldsMap[fieldName] = common.FieldUpsertResult{
			FieldName: fieldName,
			Action:    action,
		}
	}

	if len(errorMessages) != 0 {
		// Only unique errors should be surfaced.
		messages := errorMessages.List()
		sort.Strings(messages)

		return nil, fmt.Errorf("%w: %v", common.ErrBadRequest, strings.Join(messages, "; "))
	}

	return &common.UpsertMetadataResult{
		Success: true,
		Fields:  fields,
	}, nil
}

type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    Body     `xml:"Body"`
}

type Body struct {
	Response UpsertMetadataResponse `xml:"upsertMetadataResponse"`
}

type UpsertMetadataResponse struct {
	Results []UpsertMetadataResult `xml:"result"`
}

type UpsertMetadataResult struct {
	Created  bool    `xml:"created"`
	Errors   []Error `xml:"errors"`
	FullName string  `xml:"fullName"`
	Success  bool    `xml:"success"`
}

type Error struct {
	Message    string `xml:"message"`
	StatusCode string `xml:"statusCode"`
}
