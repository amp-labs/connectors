package associations

import (
	"strconv"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMakePayloadsForRelationships(t *testing.T) {

	var (
		// Relationships correspond to the numbers used by Hubspot.
		// https://developers.hubspot.com/docs/api-reference/crm-associations-v4/guide#deal-to-object
		relDealToContact           = 3
		relDealToContactStr        = strconv.Itoa(relDealToContact)
		relDealToPrimaryCompany    = 5
		relDealToPrimaryCompanyStr = strconv.Itoa(relDealToPrimaryCompany)
		relDealToCompany           = 341
		relDealToCompanyStr        = strconv.Itoa(relDealToCompany)
	)

	type record struct {
		fromObjectID string
		associations *BatchInput
	}

	type input struct {
		schema  *ObjectSchemaResponse
		records []record
	}

	tests := []struct {
		name             string
		input            input
		expectedPayloads map[objectRelationship]*BatchCreatePayload
		expectedWarning  error
	}{
		{
			name: "Empty associations",
			input: input{
				schema: &ObjectSchemaResponse{},
				records: []record{
					{
						fromObjectID: "deal_50",
						associations: &BatchInput{},
					},
				},
			},
			expectedPayloads: nil,
			expectedWarning:  nil,
		},
		{
			name: "Missing type ID in input and unknown association type ID",
			input: input{
				schema: &ObjectSchemaResponse{
					Associations: []ObjectSchemaAssociationResponse{
						{ID: relDealToContactStr, FromObjectTypeID: "0-3", ToObjectTypeID: "0-1"},
					},
				},
				records: []record{
					{
						fromObjectID: "deal_50",
						associations: &BatchInput{{
							To: Identifier{ID: "contact_71"},
							// missing type
						}, {
							To:    Identifier{ID: "unknown1"},
							Types: []Type{{ID: 999}}, // invalid
						}},
					},
				},
			},
			expectedPayloads: nil,
			expectedWarning: testutils.ExpectedOneOfErr{
				ErrMissingAssociationTypeID,
				ErrUnresolvedAssociationTypeID,
			},
		},
		{
			name: "Successful Deal with only one unknown association type ID",
			input: input{
				schema: &ObjectSchemaResponse{
					Associations: []ObjectSchemaAssociationResponse{
						{ID: relDealToContactStr, FromObjectTypeID: "0-3", ToObjectTypeID: "0-1"},
					},
				},
				records: []record{
					{
						fromObjectID: "deal_50",
						associations: &BatchInput{{
							To:    Identifier{ID: "contact_71"},
							Types: []Type{{ID: relDealToContact}}, // valid
						}, {
							To:    Identifier{ID: "unknown1"},
							Types: []Type{{ID: 999}}, // invalid
						}},
					},
				},
			},
			expectedPayloads: map[objectRelationship]*BatchCreatePayload{
				{from: "0-3", to: "0-1"}: { // deals to contacts
					Inputs: []CreateDefinition{
						{
							From:  Identifier{ID: "deal_50"},
							To:    Identifier{ID: "contact_71"},
							Types: []Type{{ID: relDealToContact}},
						},
					},
				},
			},
			expectedWarning: ErrUnresolvedAssociationTypeID,
		},
		{
			name: "Map deals to multiple contacts and companies",
			input: input{
				schema: &ObjectSchemaResponse{
					Associations: []ObjectSchemaAssociationResponse{
						{ID: relDealToContactStr, FromObjectTypeID: "0-3", ToObjectTypeID: "0-1"},
						{ID: relDealToPrimaryCompanyStr, FromObjectTypeID: "0-3", ToObjectTypeID: "0-2"},
						{ID: relDealToCompanyStr, FromObjectTypeID: "0-3", ToObjectTypeID: "0-2"},
					},
				},
				records: []record{
					{
						fromObjectID: "deal_50",
						// 3 contacts and 2 companies in various order:
						associations: &BatchInput{{
							To:    Identifier{ID: "contact_71"},
							Types: []Type{{ID: relDealToContact}},
						}, {
							To:    Identifier{ID: "contact_72"},
							Types: []Type{{ID: relDealToContact}},
						}, {
							To:    Identifier{ID: "company_61"},
							Types: []Type{{ID: relDealToPrimaryCompany}, {ID: relDealToCompany}},
						}, {
							To:    Identifier{ID: "contact_73"},
							Types: []Type{{ID: relDealToContact}},
						}, {
							To:    Identifier{ID: "company_62"},
							Types: []Type{{ID: relDealToCompany}},
						}},
					},
					{
						fromObjectID: "deal_51",
						// 1 contact and 1 company:
						associations: &BatchInput{{
							To:    Identifier{ID: "contact_21"},
							Types: []Type{{ID: relDealToContact}},
						}, {
							To:    Identifier{ID: "company_91"},
							Types: []Type{{ID: relDealToPrimaryCompany}, {ID: relDealToCompany}},
						}},
					},
					{
						fromObjectID: "deal_52",
						// 1 company:
						associations: &BatchInput{{
							To:    Identifier{ID: "company_101"},
							Types: []Type{{ID: relDealToPrimaryCompany}, {ID: relDealToCompany}},
						}},
					},
				},
			},
			// Payloads are grouped
			expectedPayloads: map[objectRelationship]*BatchCreatePayload{
				{from: "0-3", to: "0-1"}: { // deals to contacts
					Inputs: []CreateDefinition{
						{
							From:  Identifier{ID: "deal_50"},
							To:    Identifier{ID: "contact_71"},
							Types: []Type{{ID: relDealToContact}},
						}, {
							From:  Identifier{ID: "deal_50"},
							To:    Identifier{ID: "contact_72"},
							Types: []Type{{ID: relDealToContact}},
						}, {
							From:  Identifier{ID: "deal_50"},
							To:    Identifier{ID: "contact_73"},
							Types: []Type{{ID: relDealToContact}},
						}, {
							From:  Identifier{ID: "deal_51"},
							To:    Identifier{ID: "contact_21"},
							Types: []Type{{ID: relDealToContact}},
						},
					},
				},
				{from: "0-3", to: "0-2"}: { // deals to companies
					Inputs: []CreateDefinition{
						{
							From:  Identifier{ID: "deal_50"},
							To:    Identifier{ID: "company_61"},
							Types: []Type{{ID: relDealToPrimaryCompany}, {ID: relDealToCompany}},
						}, {
							From:  Identifier{ID: "deal_50"},
							To:    Identifier{ID: "company_62"},
							Types: []Type{{ID: relDealToCompany}},
						}, {
							From:  Identifier{ID: "deal_51"},
							To:    Identifier{ID: "company_91"},
							Types: []Type{{ID: relDealToPrimaryCompany}, {ID: relDealToCompany}},
						}, {
							From:  Identifier{ID: "deal_52"},
							To:    Identifier{ID: "company_101"},
							Types: []Type{{ID: relDealToPrimaryCompany}, {ID: relDealToCompany}},
						},
					},
				},
			},
			expectedWarning: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := NewBatchCreateParams("deals")
			for _, data := range tt.input.records {
				params.WithAssociation(data.fromObjectID, data.associations)
			}

			actualPayloads, actualWarning := params.makePayloadsForRelationships(tt.input.schema)

			testutils.AssertErrorIs(t, tt.expectedWarning, actualWarning, "expected errors differ")
			assert.Equal(t, len(tt.expectedPayloads), len(actualPayloads), "number of payloads differ")

			for relationship, expectedPayload := range tt.expectedPayloads {
				actualPayload, ok := actualPayloads[relationship]
				assert.True(t, ok, "missing relationship", relationship)

				assert.Equal(t, len(expectedPayload.Inputs), len(actualPayload.Inputs),
					"number of payloads differ for relationship %v", relationship)
				assert.ElementsMatch(t, expectedPayload.Inputs, actualPayload.Inputs, "payloads differ")
			}
		})
	}
}
