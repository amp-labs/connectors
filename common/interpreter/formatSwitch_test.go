package interpreter

import (
	"errors"
	"testing"
)

func TestFormatSwitchParseJSON(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name     string
		selector *FormatSwitch
		input    string
		expected error
	}{
		{
			name:     "Missing templates produces unknown format",
			selector: NewFormatSwitch(),
			input:    ``,
			expected: ErrUnknownResponseFormat,
		},
		{
			name: "Successful single template",
			selector: NewFormatSwitch(FormatTemplate{
				MustKeys: nil,
				Template: func() ErrorDescriptor { return &sampleTestFormatStatus{} },
			}),
			input:    `{"status": "bad request"}`,
			expected: errTestResStatus,
		},
		{
			name: "Format order matters",
			selector: NewFormatSwitch(FormatTemplate{
				MustKeys: []string{"code"},
				Template: func() ErrorDescriptor { return &sampleTestFormatCode{} },
			}, FormatTemplate{
				MustKeys: []string{"status"},
				Template: func() ErrorDescriptor { return &sampleTestFormatStatus{} },
			}),
			input:    `{"status": "bad request", "code": "251"}`,
			expected: errTestResCode,
		},
		{
			name: "All keys must match for template to be selected",
			selector: NewFormatSwitch(FormatTemplate{
				MustKeys: []string{"code", "messages"},
				Template: func() ErrorDescriptor { return &sampleTestFormatCode{} },
			}, FormatTemplate{
				MustKeys: []string{"status", "type"},
				Template: func() ErrorDescriptor { return &sampleTestFormatStatus{} },
			}, FormatTemplate{
				MustKeys: []string{"description"},
				Template: func() ErrorDescriptor { return &sampleTestFormatDescription{} },
			}),
			input:    `{"status": "bad request", "description": "missing required field", "code": "251"}`,
			expected: errTestResDescription,
		},
		{
			name: "No match defaults to unknown format conclusion",
			selector: NewFormatSwitch(FormatTemplate{
				MustKeys: []string{"code"},
				Template: func() ErrorDescriptor { return &sampleTestFormatCode{} },
			}, FormatTemplate{
				MustKeys: []string{"status"},
				Template: func() ErrorDescriptor { return &sampleTestFormatStatus{} },
			}, FormatTemplate{
				MustKeys: []string{"description"},
				Template: func() ErrorDescriptor { return &sampleTestFormatDescription{} },
			}),
			input:    `{}`,
			expected: ErrUnknownResponseFormat,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			descriptor := tt.selector.ParseJSON([]byte(tt.input))

			output := descriptor.CombineErr(errors.New("base-from-test")) // nolint:goerr113

			if tt.expected == nil {
				t.Fatalf("%s test is missing output expectation", tt.name)
			}

			if !errors.Is(output, tt.expected) {
				t.Fatalf("%s: expected: (%v), got: (%v)", tt.name, tt.expected, output)
			}
		})
	}
}

var (
	errTestResStatus      = errors.New("status")
	errTestResCode        = errors.New("code")
	errTestResDescription = errors.New("description")
)

type sampleTestFormatStatus struct {
	Status string `json:"status"`
}

func (f sampleTestFormatStatus) CombineErr(base error) error {
	return errTestResStatus
}

type sampleTestFormatCode struct {
	Code string `json:"code"`
}

func (f sampleTestFormatCode) CombineErr(base error) error {
	return errTestResCode
}

type sampleTestFormatDescription struct {
	Description string `json:"description"`
}

func (f sampleTestFormatDescription) CombineErr(base error) error {
	return errTestResDescription
}
