package housecallpro

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractAssociationsSkipsWhenNotRequested(t *testing.T) {
	t.Parallel()

	customer := map[string]any{
		"id":         "cus_b0f661aa89324111b575da039c45e19f",
		"first_name": "Walter",
	}

	rows := []common.ReadResultRow{{
		Id:  "job_2fad85ca2c6c43b7bdbc01f0d12ff1c4",
		Raw: map[string]any{"id": "job_2fad85ca2c6c43b7bdbc01f0d12ff1c4", "customer": customer},
	}}

	// No associations requested.
	extractAssociations("jobs", nil, rows)
	assert.Nil(t, rows[0].Associations)

	// Unknown association for jobs.
	extractAssociations("jobs", []string{"unknown"}, rows)
	assert.Nil(t, rows[0].Associations)

	// Object that has no association mappings.
	extractAssociations("customers", []string{customerObject}, rows)
	assert.Nil(t, rows[0].Associations)
}

func TestExtractAssociationsAttachesJobCustomer(t *testing.T) {
	t.Parallel()

	customer := map[string]any{
		"id":         "cus_b0f661aa89324111b575da039c45e19f",
		"first_name": "Walter",
	}

	rows := []common.ReadResultRow{{
		Id:  "job_2fad85ca2c6c43b7bdbc01f0d12ff1c4",
		Raw: map[string]any{"id": "job_2fad85ca2c6c43b7bdbc01f0d12ff1c4", "customer": customer},
	}}

	extractAssociations("jobs", []string{customerObject}, rows)

	require.Len(t, rows[0].Associations[customerObject], 1)
	assert.Equal(t, "cus_b0f661aa89324111b575da039c45e19f", rows[0].Associations[customerObject][0].ObjectId)
}

func TestAttachEmbeddedAssociation(t *testing.T) {
	t.Parallel()

	customer := map[string]any{
		"id":         "cus_b0f661aa89324111b575da039c45e19f",
		"first_name": "Walter",
		"last_name":  "Whitman",
	}

	rows := []common.ReadResultRow{{
		Id:  "job_2fad85ca2c6c43b7bdbc01f0d12ff1c4",
		Raw: map[string]any{"id": "job_2fad85ca2c6c43b7bdbc01f0d12ff1c4", "customer": customer},
	}}

	attachEmbeddedAssociation(rows, customerObject, "customer")

	require.Len(t, rows[0].Associations[customerObject], 1)
	assert.Equal(t, "cus_b0f661aa89324111b575da039c45e19f", rows[0].Associations[customerObject][0].ObjectId)
	assert.Equal(t, customer, rows[0].Associations[customerObject][0].Raw)
}

func TestAttachEmbeddedAssociationSkipsMissingOrEmpty(t *testing.T) {
	t.Parallel()

	rows := []common.ReadResultRow{
		{Raw: map[string]any{}},                             // no customer
		{Raw: map[string]any{"customer": nil}},              // null customer
		{Raw: map[string]any{"customer": map[string]any{}}}, // empty customer
		{Raw: map[string]any{"customer": map[string]any{ // customer without id
			"first_name": "Whitman",
		}}},
	}

	attachEmbeddedAssociation(rows, customerObject, "customer")

	for i := range rows {
		assert.Nil(t, rows[i].Associations, "row %d should have no associations", i)
	}
}
