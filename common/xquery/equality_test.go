package xquery

import (
	"testing"
)

func TestXMLEquals(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []struct {
		name    string
		xml1    string
		xml2    string
		want    bool
		skipNil bool // special flag for nil panics
	}{
		{
			name: "Identical XML documents",
			xml1: `<root><a>text</a></root>`,
			xml2: `<root><a>text</a></root>`,
			want: true,
		},
		{
			name: "Different attribute order but equal",
			xml1: `<root a="1" b="2"/>`,
			xml2: `<root b="2" a="1"/>`,
			want: true,
		},
		{
			name: "Attribute value mismatch",
			xml1: `<root a="1"/>`,
			xml2: `<root a="2"/>`,
			want: false,
		},
		{
			name: "Namespace mismatch",
			xml1: `<root xmlns="ns1"/>`,
			xml2: `<root xmlns="ns2"/>`,
			want: false,
		},
		{
			name: "Child element order does not matter",
			xml1: `<root><a/><b/></root>`,
			xml2: `<root><b/><a/></root>`,
			want: true,
		},
		{
			name: "Whitespace ignored in inner text",
			xml1: `<root>  text   </root>`,
			xml2: `<root>text</root>`,
			want: true,
		},
		{
			name: "Different text content",
			xml1: `<root>abc</root>`,
			xml2: `<root>xyz</root>`,
			want: false,
		},
		{
			name: "One document missing a child",
			xml1: `<root><a/></root>`,
			xml2: `<root></root>`,
			want: false,
		},
		{
			name:    "Nil receiver panic",
			xml1:    ``,
			xml2:    `<root/>`,
			skipNil: true,
		},

		// Additional complex tests:

		{
			name: "Nested reordering",
			xml1: `<root><a><b>1</b><c>2</c></a></root>`,
			xml2: `<root><a><c>2</c><b>1</b></a></root>`,
			want: true,
		},
		{
			name: "Repeated children treated as multiset",
			xml1: `<root><a/><a/><b/></root>`,
			xml2: `<root><b/><a/><a/></root>`,
			want: true,
		},
		{
			name: "Repeated children mismatch",
			xml1: `<root><a/><a/></root>`,
			xml2: `<root><a/></root>`,
			want: false,
		},
		{
			name: "Text and elements reordered – should differ",
			xml1: `<root>Hello <b>world</b></root>`,
			xml2: `<root><b>world</b> Hello</root>`,
			want: false,
		},
		{
			name: "Trimmed leaf text matches",
			xml1: `<root><msg>   Hello   </msg></root>`,
			xml2: `<root><msg>Hello</msg></root>`,
			want: true,
		},
		{
			name: "Different structure with same tokens",
			xml1: `<root><a><b/></a></root>`,
			xml2: `<root><a/><b/></root>`,
			want: false,
		},
		{
			name: "Deep reordering across branches – equal",
			xml1: `<root>
					<x>
						<a><b>1</b><c>2</c></a>
						<d/>
					</x>
					<y text="ok"/>
				   </root>`,
			xml2: `<root>
					<y text="ok"/>
					<x>
						<d/>
						<a><c>2</c><b>1</b></a>
					</x>
				   </root>`,
			want: true,
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Handle the nil-receiver test independently
			if tt.skipNil {
				second := mustParseXML(t, tt.xml2)

				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic when calling EqualsIgnoreOrder on nil receiver")
					}
				}()

				var first *XML

				first.EqualsIgnoreOrder(second)

				return
			}

			first := mustParseXML(t, tt.xml1)
			second := mustParseXML(t, tt.xml2)

			got := first.EqualsIgnoreOrder(second)
			if got != tt.want {
				t.Errorf("EqualsIgnoreOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustParseXML(t *testing.T, xml string) *XML {
	t.Helper()

	node, err := NewXML([]byte(xml))
	if err != nil {
		t.Fatalf("failed parsing XML: %v", err)
	}

	return node
}
