package interpreter

import (
	"errors"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestFormatSwitchParseJSON(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name     string
		selector *FormatSwitch
		input    string
		expected []error
	}{
		{
			name:     "Missing templates produces unknown format",
			selector: NewFormatSwitch(),
			input:    ``,
			expected: []error{ErrUnknownResponseFormat},
		},
		{
			name: "Successful single template",
			selector: NewFormatSwitch(FormatTemplate{
				MustKeys: nil,
				Template: func() ErrorDescriptor { return &sampleTestFormatStatus{} },
			}),
			input:    `{"status": "bad request"}`,
			expected: []error{errTestResStatus},
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
			expected: []error{errTestResCode},
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
			expected: []error{errTestResDescription},
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
			expected: []error{ErrUnknownResponseFormat},
		},
		{
			name: "Multiple objects are mapped to respective formats",
			selector: NewFormatSwitch(FormatTemplate{
				MustKeys: []string{"code"},
				Template: func() ErrorDescriptor { return &sampleTestFormatCode{} },
			}, FormatTemplate{
				MustKeys: []string{"status"},
				Template: func() ErrorDescriptor { return &sampleTestFormatStatus{} },
			}),
			input: `[
				{"status": "bad request", "code": "251"},
				{"random": "truly unknown format"},
				{"status": "bad request"}
			]`,
			expected: []error{ // order doesn't matter, check each exists
				errTestResCode,
				ErrUnknownResponseFormat,
				errTestResStatus,
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			descriptor := tt.selector.ParseJSON([]byte(tt.input))

			output := descriptor.CombineErr(errors.New("base-from-test")) // nolint:goerr113

			testutils.CheckErrors(t, tt.name, tt.expected, output)
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
