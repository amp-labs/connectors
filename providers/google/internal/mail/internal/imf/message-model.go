package imf

import (
	"encoding/json"
	"errors"
	"fmt"
)

// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts
//
//	{
//	 "id": string,
//	 "message": { object (Message) }
//	}
var errDraftMessageLocation = errors.New("gmail draft payload should be under 'message' property")

// newEmailMessage creates an emailMessage from a Gmail users.messages resource represented as JSON.
func newEmailMessage(gmailMessage map[string]any) (*emailMessage, error) {
	message, err := unwrapMessage(gmailMessage)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("marshal Gmail message: %w", err)
	}

	var msg emailMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal Gmail message: %w", err)
	}

	return &msg, nil
}

// Drafts have messages nested under "message" key.
// All other messages are directly positioned.
func unwrapMessage(gmailMessage map[string]any) (map[string]any, error) {
	if unwrapped, ok := gmailMessage["message"]; ok {
		// This is a draft.
		message, ok := unwrapped.(map[string]any)
		if !ok {
			return nil, errDraftMessageLocation
		}

		return message, nil
	}

	return gmailMessage, nil
}

// emailMessage is a model of Google users.messages resource.
// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages#Message
type emailMessage struct {
	Id             string                     `json:"id"`
	ThreadId       string                     `json:"threadId"`
	LabelIds       []string                   `json:"labelIds"`
	Snippet        string                     `json:"snippet"`
	HistoryId      string                     `json:"historyId"`
	InternalDate   string                     `json:"internalDate"`
	Payload        emailMessagePart           `json:"payload"`
	SizeEstimate   int                        `json:"sizeEstimate"`
	Raw            string                     `json:"raw"`
	ClassLabelVals []classificationLabelValue `json:"classificationLabelValues"`
}

type emailMessagePart struct {
	// The immutable ID of the message part.
	PartId string `json:"partId"`
	// The MIME type of the message part.
	MimeType string `json:"mimeType"`
	// The filename of the attachment. Only present if this message part represents an attachment.
	Filename string `json:"filename"`
	// List of headers on this message part. For the top-level message part,
	// representing the entire message payload, it will contain
	// the standard RFC 2822 email headers such as To, From, and Subject.
	Headers []emailHeader `json:"headers"`
	// The message part body for this part, which may be empty for container MIME message parts.
	Body messagePartBody `json:"body"`
	// The child MIME message parts of this part. This only applies to container MIME message parts,
	// for example multipart/*. For non-container MIME message part types,
	// such as text/plain, this field is empty. For more information, see RFC 1521.
	Parts []emailMessagePart `json:"parts"`
}

type emailHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type messagePartBody struct {
	AttachmentId string `json:"attachmentId"`
	Size         int    `json:"size"`
	Data         string `json:"data"`
}

type classificationLabelValue struct {
	LabelId []string                        `json:"labelId"`
	Fields  []classificationLabelFieldValue `json:"fields"`
}

type classificationLabelFieldValue struct {
	FieldId   string `json:"fieldId"`
	Selection string `json:"selection"`
}
