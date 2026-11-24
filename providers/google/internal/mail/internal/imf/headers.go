package imf

import (
	"net/textproto"
	"strings"
)

// Build the MIMEHeaders map from all payload headers.
//
// Headers: all RFC 2822 headers from message.Payload.Headers mapped into
// a textproto.MIMEHeader (keyed by canonical header name).
func (m emailMessage) getMimeHeaders() textproto.MIMEHeader {
	headers := make(textproto.MIMEHeader)
	for _, h := range m.Payload.Headers {
		headers.Add(h.Name, h.Value)
	}

	return headers
}

func (p emailMessagePart) getHeader(name string) string {
	for _, h := range p.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}

	return ""
}
