package objects_test

import (
	"testing"

	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot"
	"github.com/amp-labs/connectors/providers/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMatrixLoads is the smoke test for the generated artifact. A matrix that does not
// parse, or that has silently emptied itself, is worse than no matrix -- every lookup would
// answer "unknown" and nobody would notice.
func TestMatrixLoads(t *testing.T) {
	t.Parallel()

	matrix, err := objects.Load()
	require.NoError(t, err)
	assert.NotEmpty(t, matrix.Timestamp)
	assert.Greater(t, len(matrix.Providers), 50,
		"the matrix should cover the providers that ship schemas, not a handful")
}

// TestActivityObjectsAreRecorded is the question this whole artifact exists to answer.
//
// These are the objects an engagement platform syncs -- calls, emails, meetings, messages --
// and before this there was no artifact that said whether a provider handled them. Each was
// discoverable only by standing up a connection and trying.
func TestActivityObjectsAreRecorded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider string
		module   string
		object   string
	}{
		{providers.Salesforce, "", "Task"},
		{providers.Salesforce, "", "EmailMessage"},
		{providers.Salesforce, "", "VoiceCall"},
		{providers.Salesforce, "", "Event"},
		{providers.SalesforceJWT, "", "Task"},
		// The object behind HubSpot SMS, WhatsApp and LinkedIn messages, and the one
		// most often missed.
		{providers.Hubspot, "crm", "communications"},
		{providers.Hubspot, "crm", "calls"},
		{providers.Hubspot, "crm", "emails"},
		{providers.Hubspot, "crm", "meetings"},
		{providers.DynamicsCRM, "", "phonecalls"},
		{providers.DynamicsCRM, "", "tasks"},
	}

	for _, testCase := range tests {
		t.Run(testCase.provider+"/"+testCase.object, func(t *testing.T) {
			t.Parallel()

			read := objects.Supports(testCase.provider, testCase.module, testCase.object, objects.OperationRead)
			assert.Equal(t, objects.AnswerSupported, read, "read should be supported")

			write := objects.Supports(testCase.provider, testCase.module, testCase.object, objects.OperationWrite)
			assert.Equal(t, objects.AnswerSupported, write, "write should be supported")
		})
	}
}

// TestHubspotActivityObjectsMatchTheConnector keeps the hand-declared list honest.
//
// hubspot.KnownObjectTypes is keyed to HubSpot's published object type IDs, so it is the
// grounded list. Declaring an object here that the connector does not know about would be
// asserting support for something nothing can address.
func TestHubspotActivityObjectsMatchTheConnector(t *testing.T) {
	t.Parallel()

	known := make(map[string]bool, len(hubspot.KnownObjectTypes))
	for _, name := range hubspot.KnownObjectTypes {
		known[name] = true
	}

	declared := objects.DeclaredObjects(providers.Hubspot)
	require.NotEmpty(t, declared)

	for module, names := range declared {
		for _, name := range names {
			assert.True(t, known[name],
				"declared object %q (module %q) is not in hubspot.KnownObjectTypes", name, module)
		}
	}
}

// TestObjectLookupIsCaseInsensitive. Manifests carry the provider's casing while shipped
// schemas carry whatever the provider's API returns, and a case mismatch answered as
// "unsupported" is a wrong answer rather than a missing one -- exactly what an
// authoritative matrix must not produce.
func TestObjectLookupIsCaseInsensitive(t *testing.T) {
	t.Parallel()

	for _, spelling := range []string{"Task", "task", "TASK"} {
		assert.Equal(t, objects.AnswerSupported,
			objects.Supports(providers.Salesforce, "", spelling, objects.OperationRead),
			"spelling %q should resolve", spelling)
	}
}

// TestUnknownIsNotUnsupported is the distinction the Answer type exists for.
//
// A generic provider reaches whatever the credential can, so an unlisted object may well
// work; saying "unsupported" there would send someone away from something that would have
// succeeded. A provider we hold no data on at all is equally unknown.
func TestUnknownIsNotUnsupported(t *testing.T) {
	t.Parallel()

	t.Run("an unlisted object on a generic provider is unknown", func(t *testing.T) {
		t.Parallel()

		entry, ok := objects.ForProvider(providers.Salesforce)
		require.True(t, ok)
		require.True(t, entry.Generic, "salesforce is object-generic")

		assert.Equal(t, objects.AnswerUnknown,
			objects.Supports(providers.Salesforce, "", "Some_Custom_Object__c", objects.OperationRead))
	})

	t.Run("a provider not in the matrix is unknown", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, objects.AnswerUnknown,
			objects.Supports("a-provider-that-does-not-exist", "", "anything", objects.OperationRead))
	})
}

// TestFixedListProviderReportsUnsupported: for a provider whose objects are enumerated from
// its own shipped schema, absence is a real answer rather than a gap.
func TestFixedListProviderReportsUnsupported(t *testing.T) {
	t.Parallel()

	matrix, err := objects.Load()
	require.NoError(t, err)

	var fixed string

	for name, entry := range matrix.Providers {
		if !entry.Generic && len(entry.Objects) > 0 {
			fixed = name

			break
		}
	}

	require.NotEmpty(t, fixed, "expected at least one non-generic provider in the matrix")

	assert.Equal(t, objects.AnswerUnsupported,
		objects.Supports(fixed, "", "an_object_this_provider_does_not_have", objects.OperationRead))
}

// TestSourceIsRecorded lets a caller tell a fact from a claim: schema entries are generated
// from the provider's own API, declared ones were written by a human.
func TestSourceIsRecorded(t *testing.T) {
	t.Parallel()

	entry, ok := objects.ForProvider(providers.Salesforce)
	require.True(t, ok)

	task, found := entry.Objects[""]["Task"]
	require.True(t, found)
	assert.Equal(t, objects.SourceDeclared, task.Source)
}
