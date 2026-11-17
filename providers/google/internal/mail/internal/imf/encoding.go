// Package imf bridges Google's Gmail message format with the Internet
// message Format (IMF).
//
// It wraps github.com/jordan-wright/email and converts a Gmail message
// resource into an RFC 2822-compliant MIME message:
//
//	https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages
//
// The conversion reconstructs the MIME message from Gmail's message record.
// It preserves the semantic content of the email-headers, MIME structure,
// bodies, attachments, and inline content, but does not necessarily reproduce
// Gmail's original raw MIME serialization byte for byte.
package imf

import (
	"github.com/jordan-wright/email"
)

// FormatMessage converts a Gmail API message representation into an
// Internet Message Format (IMF) message.
//
// It accepts the JSON-shaped map returned by the Gmail API for a Message resource:
//
//	https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages
//
// The returned text is either:
//
//   - the original raw MIME message, already base64url encoded, when
//     gmailMessage contains a non-empty "raw" field; or
//   - a newly constructed IMF/MIME message, not encoded, built from the
//     message headers, body parts, attachments, and inline images.
//
// isEncoded reports whether the returned text is already base64url encoded.
// It is true only for the raw-message fast path. In all other cases, the
// caller must encode the returned IMF bytes before supplying them to the
// Gmail API.
//
// The function preserves inline-image semantics when reconstructing a
// message. An image referenced by the HTML body through a cid: URL is added
// as a related MIME part with its original Content-ID and an inline
// Content-Disposition, allowing mail clients to render it in the message body
// rather than present it as a downloadable attachment.
//
// Gmail-only metadata, such as id, threadId, labelIds, snippet, historyId,
// internalDate, and sizeEstimate, is not part of IMF and is therefore not
// included in the generated message.
//
// Example:
//
//	text, encoded, err := imf.FormatMessage(gmailMessage)
//	if err != nil {
//		return err
//	}
//
//	if !encoded {
//		text = base64.URLEncoding.EncodeToString(text)
//	}
//
//	msg := gmail.Message{
//		Raw: string(text),
//	}
//
// If gmailMessage["raw"] is present and non-empty, it is returned unchanged
// after conversion to []byte; no headers, body, attachments, or inline images
// are parsed or rebuilt.
func FormatMessage(gmailMessage map[string]any) (text []byte, isEncoded bool, err error) {
	message, err := newEmailMessage(gmailMessage)
	if err != nil {
		return nil, false, err
	}

	// A raw message already contains the complete IMF representation,
	// including headers, body, attachments, and inline images.
	if message.Raw != "" {
		return []byte(message.Raw), true, nil
	}

	textBody, htmlBody, err := message.extractBody()
	if err != nil {
		return nil, false, err
	}

	attachments, err := message.extractAttachments(htmlBody)
	if err != nil {
		return nil, false, err
	}

	internetMsg := email.Email{
		// Plain-text body decoded from the first text/plain part,
		// could be nested within multipart containers.
		Text: textBody,
		// HTML body decoded from the first text/html part,
		// could be nested within multipart containers.
		HTML: htmlBody,
		// Preserve the original RFC 2822 headers from the Gmail message.
		Headers: message.getMimeHeaders(),
	}

	for _, attachment := range attachments {
		if err = attachment.Apply(&internetMsg); err != nil {
			return nil, false, err
		}
	}

	text, err = internetMsg.Bytes()

	return text, false, err
}
