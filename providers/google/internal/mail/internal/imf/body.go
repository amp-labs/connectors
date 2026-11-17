package imf

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Extract body content from payload parts.
// For text/plain parts, decode base64 data into Text.
// For text/html parts, decode base64 data into HTML.
// For multipart, recurse through message.Payload.Parts.
func (m emailMessage) extractBody() (text, html []byte, err error) {
	return extractBody(m.Payload)
}

func extractBody(part emailMessagePart) (text, html []byte, err error) { // nolint:cyclop
	switch {
	case part.Filename == "" && part.MimeType == "text/plain":
		text, err = decodeBodyData(part.Body.Data)
		if err != nil {
			return nil, nil, err
		}
	case part.Filename == "" && part.MimeType == "text/html":
		html, err = decodeBodyData(part.Body.Data)
		if err != nil {
			return nil, nil, err
		}
	case strings.HasPrefix(part.MimeType, "multipart/"):
		for _, child := range part.Parts {
			t, h, err := extractBody(child)
			if err != nil {
				return nil, nil, err
			}

			if text == nil {
				text = t // Out of all parts the "text/plain" was found.
			}

			if html == nil {
				html = h // Out of all parts the "text/html" was found.
			}
		}
	default:
		return nil, nil, nil
	}

	return text, html, nil
}

func decodeBodyData(data string) ([]byte, error) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(data)
	}

	if err != nil {
		return nil, fmt.Errorf("decode base64url data: %w", err)
	}

	return decoded, nil
}
