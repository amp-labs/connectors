package xquery

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestXMLSetVariousData(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	dataFile := testutils.DataFromFile(t, "envelope.xml")
	changeTextSession := testutils.DataFromFile(t, "change-text-session.xml")
	changeNodeNameField := testutils.DataFromFile(t, "change-node-name-field.xml")
	changeNodeAddList := testutils.DataFromFile(t, "change-node-add-list.xml")

	tests := []struct {
		name     string
		modify   func(node *XML) error
		expected []byte
	}{
		{
			name: "Successfully modify text",
			modify: func(node *XML) error {
				node.FindOne("//sessionId").GetChild().
					SetDataText("Modified Session From Test")

				return nil
			},
			expected: changeTextSession,
		},
		{
			name: "Cannot set node to text",
			modify: func(node *XML) error {
				tree, err := NewXML([]byte(`<tree>Plumb</tree>`))
				if err != nil {
					return err
				}
				node.FindOne("//sessionId").GetChild().
					SetDataNode(tree)

				return nil
			},
			expected: dataFile, // no changes
		},
		{
			name: "Replace all node children with one",
			modify: func(node *XML) error {
				tree, err := NewXML([]byte(`<tree>Pomegranate</tree>`))
				if err != nil {
					return err
				}
				node.FindOne("//nameField").
					SetDataNode(tree)

				return nil
			},
			expected: changeNodeNameField,
		},
		{
			name: "Cannot replace element node with text",
			modify: func(node *XML) error {
				node.FindOne("//nameField").GetChild().
					SetDataText("Trying to set text on invalid target")

				return nil
			},
			expected: dataFile, // no changes
		},
		{
			name: "All children are overridden onto target node",
			modify: func(node *XML) error {
				// this xml has many children and only the first NodeElement kind
				// will be used for setting
				tree, err := NewXML([]byte(`
					<first>The first element tag for a parent</first>
					<!-- This is a verbose comment -->
					<second>Another element</second>
					<!-- Middle comment -->
					<third>Last element</third>
				`))
				if err != nil {
					return err
				}
				node.FindOne("//createMetadata").
					SetDataNode(tree)

				return nil
			},
			expected: changeNodeAddList,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := NewXML(dataFile)
			if err != nil {
				t.Fatalf("failed to start test, input file is not XML")
			}

			err = tt.modify(data) // applies modifications unique to the test
			if err != nil {
				t.Fatalf("failed to test scenario during xml modification")
			}

			output := data.RawXML()

			// convert expected data to the same format
			afterModifications, err := NewXML(tt.expected)
			if err != nil {
				t.Fatalf("failed test scenario, expectation is not XML")
			}

			expected := afterModifications.RawXML()

			if !reflect.DeepEqual(output, expected) {
				diff := deep.Equal(output, expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)",
					tt.name, expected, output, diff)
			}
		})
	}
}
