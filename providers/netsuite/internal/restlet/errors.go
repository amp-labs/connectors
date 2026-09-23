package restlet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRestletError              = errors.New("restlet error")
	ErrUnsupportedFilterOperator = errors.New("unsupported filter operator")
)

const (
	// maxStackFrames caps how many stack frames are appended to the error message.
	maxStackFrames = 8
	// maxDetailsLen caps the length of non-stack errorDetails appended to the error message.
	maxDetailsLen = 1000
)

func parseRestletError(resp *restletResponse) error {
	var errBody restletErrorBody
	if err := json.Unmarshal(resp.Body, &errBody); err != nil {
		return fmt.Errorf("%w: status=%s", ErrRestletError, resp.Header.Status)
	}

	if details := formatErrorDetails(errBody.ErrorDetails); details != "" {
		return fmt.Errorf("%w: [%s] %s (%s)", ErrRestletError, errBody.ErrorCode, errBody.ErrorMessage, details)
	}

	return fmt.Errorf("%w: [%s] %s", ErrRestletError, errBody.ErrorCode, errBody.ErrorMessage)
}

// formatErrorDetails renders the RESTlet's errorDetails for inclusion in an error message.
// Generic NetSuite errors such as UNEXPECTED_ERROR carry nothing useful in errorMessage;
// the stack in errorDetails is the only pointer to where the failure happened.
// Returns "" when there are no details.
func formatErrorDetails(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) || bytes.Equal(raw, []byte("{}")) {
		return ""
	}

	var details struct {
		Stack json.RawMessage `json:"stack"`
	}

	if err := json.Unmarshal(raw, &details); err == nil && len(details.Stack) > 0 {
		if frames := stackFrames(details.Stack); len(frames) > 0 {
			return "stack: " + strings.Join(frames, "; ")
		}
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return ""
	}

	return "details: " + truncate(compact.String(), maxDetailsLen)
}

// stackFrames extracts the "at ..." frames from a SuiteScript stack, which the RESTlet
// returns either as a string or as an array of multi-line strings.
func stackFrames(raw json.RawMessage) []string {
	var parts []string

	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		parts = []string{single}
	} else if err := json.Unmarshal(raw, &parts); err != nil {
		return nil
	}

	var frames []string

	for _, part := range parts {
		for line := range strings.SplitSeq(part, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "at ") {
				continue
			}

			frames = append(frames, line)
			if len(frames) == maxStackFrames {
				return frames
			}
		}
	}

	return frames
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n] + "..."
}
