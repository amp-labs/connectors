package imf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jordan-wright/email"
)

type emailAttachment struct {
	Data               []byte
	Filename           string
	MimeType           string
	ContentID          string
	ContentDisposition string
	IsInline           bool
}

// Extract attachments from payload parts.
// A part is an attachment if it has a non-empty Filename and body Data.
func (m emailMessage) extractAttachments(htmlBody []byte) ([]emailAttachment, error) {
	return extractAttachments(m.Payload, htmlBody)
}

func extractAttachments(part emailMessagePart, htmlBody []byte) ([]emailAttachment, error) {
	var attachments []emailAttachment

	if part.Filename != "" && part.Body.Data != "" {
		data, err := decodeBodyData(part.Body.Data)
		if err != nil {
			return nil, fmt.Errorf("decode attachment %q: %w", part.Filename, err)
		}

		disposition := part.getHeader("Content-Disposition")
		contentId := part.getHeader("Content-ID")
		isInline := isInlineAttachment(disposition, contentId, htmlBody)

		attachments = append(attachments, emailAttachment{
			Data:               data,
			Filename:           part.Filename,
			MimeType:           part.MimeType,
			ContentID:          contentId,
			ContentDisposition: disposition,
			IsInline:           isInline,
		})
	}

	for _, child := range part.Parts {
		childAttachments, err := extractAttachments(child, htmlBody)
		if err != nil {
			return nil, err
		}

		attachments = append(attachments, childAttachments...)
	}

	return attachments, nil
}

func isInlineAttachment(disposition, contentID string, htmlBody []byte) bool {
	if strings.Contains(strings.ToLower(disposition), "inline") {
		return true
	}

	if contentID == "" {
		return false
	}

	contentID = strings.Trim(contentID, "<>")

	return bytes.Contains(htmlBody, []byte("cid:"+contentID))
}

func (a emailAttachment) Apply(internetMsg *email.Email) error {
	attachment, err := internetMsg.Attach(
		bytes.NewReader(a.Data),
		a.Filename,
		a.MimeType,
	)
	if err != nil {
		return err
	}

	if a.IsInline {
		// Example:
		// 	Content-Disposition: inline; filename="house.png"
		// 	Content-Id: <ii_mwfmg0ju0>

		// Put the attachment into the multipart/related section.
		attachment.HTMLRelated = true

		// Preserve Gmail's original cid reference.
		cid := strings.Trim(a.ContentID, "<>")
		attachment.Header.Set("Content-ID", fmt.Sprintf("<%s>", cid))

		// Force inline rather than attachment.
		attachment.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", a.Filename))
	}

	return nil
}
