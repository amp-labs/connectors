package common

import (
	"fmt"
)

type API string

const (
	Salesforce API = "salesforce"
)

// ReadConfig defines what API to read from, and what we are reading.
type ReadConfig struct {
	API API
	ObjectName string
	Fields [] string
	AccessToken string
	// WorkspaceID is the ID of the workspace, subdomain, etc. that we are reading from.
	WorkspaceID string
}

type Result struct {
	// Rows is the number of total rows in the result.
	Rows int
	// Data is a list of maps, where each map represents a record that we read.
	Data [] map[string]interface{}
}

type ErrorWithStatus struct {
	// StatusCode is the HTTP status.
	StatusCode int
	// A human-readable error message.
	Message string
}

func (r ErrorWithStatus) Error() string {
	return fmt.Sprintf("status %d: message %v", r.StatusCode, r.Message)
}
