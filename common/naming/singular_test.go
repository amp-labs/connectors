package naming

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-test/deep"
)

func TestSingularString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		singular string
		plural   string
	}{
		{
			name:     "Singular input",
			input:    "admin",
			singular: "admin",
			plural:   "admins",
		},
		{
			name:     "Plural input",
			input:    "subscription_types",
			singular: "subscription_type",
			plural:   "subscription_types",
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			singular := NewSingularString(tt.input)
			singular = singular.Plural().Singular() // tautology
			output := singular.String()

			if output != tt.singular {
				failedExpectation(t, tt.name, "to_singular", tt.singular, output)
			}

			output = singular.Plural().String()
			if output != tt.plural {
				failedExpectation(t, tt.name, "to_plural", tt.plural, output)
			}
		})
	}
}

func TestSingularMarshal(t *testing.T) { // nolint:funlen
	t.Parallel()

	type RegistryS struct {
		Title SingularString            `json:"title"` // test encoding as value
		Data  map[SingularString]string `json:"data"`  // struct can be key
		Meta  map[string]string         `json:"meta"`
		Extra SingularString            `json:"extra,omitempty"` // can be empty
	}

	type RegistryP struct {
		Title PluralString            `json:"title"` // test encoding as value
		Data  map[PluralString]string `json:"data"`  // struct can be key
		Meta  map[string]string       `json:"meta"`
		Extra PluralString            `json:"extra,omitempty"` // can be empty
	}

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Singular, Marshal and unmarshal",
			run: func(t *testing.T) {
				t.Helper()

				singular := NewSingularString("potatoes")

				source := RegistryS{
					Title: singular,
					Data:  map[SingularString]string{singular: "McDonald's"},
					Meta:  map[string]string{"coca cola": "normal"},
				}

				var target RegistryS
				marshalBackAndForth(t, source, &target)
				if !reflect.DeepEqual(source, target) {
					diff := deep.Equal(source, target)
					t.Fatalf("%s:, marshalling data mismatching \ndiff: (%v)", t.Name(), diff)
				}
			},
		},
		{
			name: "Plural, Marshal and unmarshal",
			run: func(t *testing.T) {
				t.Helper()

				plural := NewPluralString("company")

				source := RegistryP{
					Title: plural,
					Data:  map[PluralString]string{plural: "Enterprise"},
					Meta:  map[string]string{"location": "California"},
				}

				var target RegistryP
				marshalBackAndForth(t, source, &target)
				if !reflect.DeepEqual(source, target) {
					diff := deep.Equal(source, target)
					t.Fatalf("%s:, marshalling data mismatching \ndiff: (%v)", t.Name(), diff)
				}
			},
		},
		{
			name: "Convert to Singular format from byte data",
			run: func(t *testing.T) {
				t.Helper()

				pluralData := `{"title":"potatoes","data":{"potatoes":"McDonald's"}}`
				var registry RegistryS
				err := json.Unmarshal([]byte(pluralData), &registry)
				check(t, err)

				_, found := registry.Data[NewSingularString("potato")]
				titleMatch := registry.Title.String() == "potato"
				if !titleMatch || !found {
					t.Fatalf("%s format was not applied from byte data, %v", t.Name(), registry)
				}
			},
		},
		{
			name: "Convert to Plural format from byte data",
			run: func(t *testing.T) {
				t.Helper()

				singularData := `{"title":"company","data":{"company":"Enterprise"}}`
				var registry RegistryP
				err := json.Unmarshal([]byte(singularData), &registry)
				check(t, err)

				_, found := registry.Data[NewPluralString("companies")]
				titleMatch := registry.Title.String() == "companies"
				if !titleMatch || !found {
					t.Fatalf("%s format was not applied from byte data, %v", t.Name(), registry)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.run(t)
		})
	}
}

func marshalBackAndForth(t *testing.T, from, to any) {
	t.Helper()

	data, err := json.Marshal(from)
	check(t, err)
	err = json.Unmarshal(data, &to)
	check(t, err)
}

func check(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: failed because: %v", t.Name(), err)
	}
}

func failedExpectation(t *testing.T, name, descr, expected, got string) {
	t.Helper()

	t.Fatalf("%s: [%v] expected: (%v), got: (%v)", name, descr, expected, got)
}
