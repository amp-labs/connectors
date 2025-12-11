package custom

import (
	"fmt"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Converts XML output from Salesforce.
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
func transformResponseToResult(resp *UpsertMetadataResponse) (*common.UpsertMetadataResult, error) {
	errorMessages := datautils.NewStringSet()
	fields := make(map[string]map[string]common.FieldUpsertResult)

	for _, result := range resp.Response.Results {
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

type UpsertMetadataResponse struct {
	Response struct {
		Results []UpsertMetadataResult `xml:"result"`
	} `xml:"upsertMetadataResponse"`
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

type ReadMetadataBody[R any] struct {
	Response ReadMetadataResponse[R] `xml:"readMetadataResponse"`
}

type ReadMetadataResponse[R any] struct {
	Results []ReadMetadataResult[R] `xml:"result"`
}

type ReadMetadataResult[R any] struct {
	Records []R `xml:"records"`
}

type PermissionSetResponse ReadMetadataBody[PermissionSet]

// GetFieldPermissions returns all FieldPermissions from the PermissionSetResponse
// that belong to the default Ampersand-managed permission set
// (see DefaultPermissionSetName).
//
// Records marked as xsi:nil="true" are skipped.
// Only records with xsi:type="PermissionSet" are processed.
//
// Example of empty response:
//
//	<readMetadataResponse>
//	  <result>
//	    <records xsi:nil="true"/>
//	  </result>
//	</readMetadataResponse>
func (r PermissionSetResponse) GetFieldPermissions() FieldPermissions {
	fieldPermissions := make(FieldPermissions)

	for _, result := range r.Response.Results {
		for _, record := range result.Records {
			if record.IsNil || record.XSIType != PermissionSetType {
				continue
			}

			if record.FullName == DefaultPermissionSetName {
				for _, permission := range record.FieldPermissions {
					fieldPermissions[permission.FullName] = permission
				}
			}
		}
	}

	return fieldPermissions
}

type PermissionSet struct {
	XSIType          string            `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	IsNil            bool              `xml:"http://www.w3.org/2001/XMLSchema-instance nil,attr"`
	FullName         string            `xml:"fullName"`
	FieldPermissions []FieldPermission `xml:"fieldPermissions"`
}

type FieldPermissions map[string]FieldPermission

// FieldPermission represents the access rights for a single field.
// This type is used both in request payloads and in responses, as the structure
// is identical in both cases.
type FieldPermission struct {
	FullName string `xml:"field"`
	Readable bool   `xml:"readable"`
	Editable bool   `xml:"editable"`
}
