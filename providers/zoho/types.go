package zoho

import (
	"fmt"
	"strings"
	"time"
)

//nolint:tagliatelle
type SubscriptionRequest struct {
	UniqueRef       string         `json:"unique_ref"         validate:"required"`
	WebhookEndPoint string         `json:"webhook_end_point"  validate:"required"`
	Duration        *time.Duration `json:"duration,omitempty"`
}

type Result struct {
	Watch []WatchResult `json:"watch"`
}

// ModuleEvent represents a module and operation combination.
type ModuleEvent string

// ModuleAPI returns the formatted string representation of the module event.
func (me ModuleEvent) ModuleAPI() (string, error) {
	parts, err := me.parts()
	if err != nil {
		return "", err
	}

	return parts[0], nil
}

func (me ModuleEvent) Operation() (string, error) {
	parts, err := me.parts()
	if err != nil {
		return "", err
	}

	return parts[1], nil
}

func (me ModuleEvent) parts() ([]string, error) {
	parts := strings.Split(string(me), ".")
	//nolint:mnd
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: %s", errInvalidModuleEvent, me)
	}

	return parts, nil
}

type SubscriptionPayload struct {
	Watch []Watch `json:"watch"`
}

//nolint:tagliatelle
type NotificationCondition struct {
	Type           string         `json:"type"`
	Module         Module         `json:"module"`
	FieldSelection FieldSelection `json:"field_selection"`
}

//nolint:tagliatelle
type Module struct {
	APIName string `json:"api_name"`
	Id      string `json:"id"`
}

type GroupOperator string

const (
	GroupOperatorOr  = "or"
	GroupOperatorAnd = "and"
)

//nolint:tagliatelle
type FieldSelection struct {
	GroupOperator GroupOperator `json:"group_operator,omitempty"`
	Group         []FieldGroup  `json:"group,omitempty"`
	Field         *Field        `json:"field,omitempty"`
}

//nolint:tagliatelle
type FieldGroup struct {
	Field         *Field       `json:"field,omitempty"`
	GroupOperator string       `json:"group_operator,omitempty"`
	Group         []FieldGroup `json:"group,omitempty"`
}

//nolint:tagliatelle
type Field struct {
	APIName string `json:"api_name"`
	ID      string `json:"id"`
}

//nolint:tagliatelle
type Watch struct {
	// ChannelID String representation of int64. Accepts negative values as well.
	ChannelID                 string                  `json:"channel_id"`
	Events                    []ModuleEvent           `json:"events"`
	NotificationCondition     []NotificationCondition `json:"notification_condition,omitempty"`
	ChannelExpiry             string                  `json:"channel_expiry,omitempty"`
	Token                     string                  `json:"token,omitempty"`
	ReturnAffectedFieldValues bool                    `json:"return_affected_field_values,omitempty"`
	NotifyURL                 string                  `json:"notify_url"`
	NotifyOnRelatedAction     bool                    `json:"notify_on_related_action,omitempty"`
}

// WatchResponse represents the top-level response from the Zoho CRM watch API.
type WatchResponse struct {
	Watch []WatchResult `json:"watch"`
}

// WatchResult represents a single watch subscription result.
type WatchResult struct {
	Code    string       `json:"code"`
	Details WatchDetails `json:"details"`
	Message string       `json:"message"`
	Status  string       `json:"status"`
}

// WatchDetails contains the details of the watch subscription.
type WatchDetails struct {
	Events []WatchEvent `json:"events"`
}

// WatchEvent represents a single event in the watch subscription.
//
//nolint:tagliatelle
type WatchEvent struct {
	ChannelExpiry string `json:"channel_expiry"`
	ResourceURI   string `json:"resource_uri"`
	ResourceID    string `json:"resource_id"`
	ResourceName  string `json:"resource_name"`
	ChannelID     string `json:"channel_id"`
}
