package deepmock

import (
	"context"

	"github.com/amp-labs/connectors/deepmock"
)

// Sample CRM-style schemas for demonstration purposes

var contactSchemaJSON = []byte(`{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"type": "object",
	"properties": {
		"id": {
			"type": "string",
			"x-amp-id-field": true,
			"description": "Unique identifier for the contact"
		},
		"email": {
			"type": "string",
			"format": "email",
			"description": "Contact's email address"
		},
		"firstName": {
			"type": "string",
			"minLength": 1,
			"description": "Contact's first name"
		},
		"lastName": {
			"type": "string",
			"minLength": 1,
			"description": "Contact's last name"
		},
		"phone": {
			"type": "string",
			"format": "phone",
			"description": "Contact's phone number"
		},
		"status": {
			"type": "string",
			"enum": ["active", "inactive"],
			"description": "Contact status"
		},
		"createdAt": {
			"type": "integer",
			"x-amp-updated-field": true,
			"description": "Timestamp when the contact was created"
		},
		"tags": {
			"type": "array",
			"items": {
				"type": "string"
			},
			"uniqueItems": true,
			"description": "Tags associated with the contact"
		}
	},
	"required": ["email"]
}`)

var companySchemaJSON = []byte(`{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"type": "object",
	"properties": {
		"id": {
			"type": "integer",
			"x-amp-id-field": true,
			"description": "Unique identifier for the company"
		},
		"name": {
			"type": "string",
			"minLength": 1,
			"description": "Company name"
		},
		"industry": {
			"type": "string",
			"enum": ["technology", "finance", "healthcare", "retail", "manufacturing"],
			"description": "Industry sector"
		},
		"employeeCount": {
			"type": "integer",
			"minimum": 1,
			"description": "Number of employees"
		},
		"website": {
			"type": "string",
			"format": "uri",
			"description": "Company website URL"
		},
		"updatedAt": {
			"type": "string",
			"format": "date-time",
			"x-amp-updated-field": true,
			"description": "Timestamp when the company was last updated"
		}
	},
	"required": ["name"]
}`)

var dealSchemaJSON = []byte(`{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"type": "object",
	"properties": {
		"id": {
			"type": "string",
			"x-amp-id-field": true,
			"description": "Unique identifier for the deal"
		},
		"title": {
			"type": "string",
			"minLength": 1,
			"description": "Deal title"
		},
		"amount": {
			"type": "number",
			"minimum": 0,
			"description": "Deal amount in dollars"
		},
		"stage": {
			"type": "string",
			"enum": ["prospecting", "qualification", "proposal", "negotiation", "closed-won", "closed-lost"],
			"description": "Current stage of the deal"
		},
		"contactId": {
			"type": "string",
			"description": "ID of the associated contact"
		},
		"companyId": {
			"type": "integer",
			"description": "ID of the associated company"
		},
		"closeDate": {
			"type": "string",
			"format": "date",
			"description": "Expected or actual close date"
		},
		"lastModified": {
			"type": "integer",
			"x-amp-updated-field": true,
			"description": "Timestamp when the deal was last modified"
		}
	},
	"required": ["title"]
}`)

// GetDeepMockConnector returns a configured deepmock connector with sample CRM schemas.
// This connector can be used for testing and demonstration purposes.
//
// The connector includes three object types:
//   - contacts: Contact records with email, name, phone, status, and tags
//   - companies: Company records with name, industry, employee count, and website
//   - deals: Deal records with title, amount, stage, and relationships to contacts/companies
//
// Example usage:
//
//	conn := GetDeepMockConnector(ctx)
//	record := conn.GenerateRandomRecord("contacts")
//	result, err := conn.Write(ctx, common.WriteParams{
//	    ObjectName: "contacts",
//	    RecordData: record,
//	})
func GetDeepMockConnector(ctx context.Context) *deepmock.Connector {
	schemas := map[string][]byte{
		"contacts":  contactSchemaJSON,
		"companies": companySchemaJSON,
		"deals":     dealSchemaJSON,
	}

	conn, err := deepmock.NewConnector(schemas)
	if err != nil {
		panic(err)
	}

	return conn
}
