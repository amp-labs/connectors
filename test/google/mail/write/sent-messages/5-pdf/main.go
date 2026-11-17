package main

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/google/mail/internal/files"
	"github.com/amp-labs/connectors/test/google/mail/internal/mailbox"
)

//go:embed email.json
var email []byte

func main() {
	flags := mailbox.ParseFlags("to", "from")

	text := `Hello,

This is MIME test 5.

This message has a PDF attached as a separate downloadable file.

Please inspect the raw message and look for:

- multipart/mixed
- text/plain and/or text/html
- application/pdf
- Content-Disposition: attachment
- Content-Transfer-Encoding: base64
- filename="mime-test.pdf"

Regards,
MIME Explorer
`

	html := `<div dir="ltr"><div dir="ltr"><pre>Hello,

This is MIME test 5.

This message has a PDF attached as a separate downloadable file.

Please inspect the raw message and look for:

- multipart/mixed
- text/plain and/or text/html
- application/pdf
- Content-Disposition: attachment
- Content-Transfer-Encoding: base64
- filename=&quot;mime-test.pdf&quot;

Regards,
MIME Explorer</pre></div>
</div>
`

	payload, err := mailbox.CreatePayload(email,
		mailbox.NewVar("To", flags.To),
		mailbox.NewVar("From", flags.From),
		mailbox.NewVar("Subject", "{005} Email with PDF attachment"),
		mailbox.NewEncodedVar("TEXT", []byte(text)),
		mailbox.NewEncodedVar("HTML", []byte(html)),
		mailbox.NewEncodedVar("Document", files.DocumentAboutGoogleDev),
	)
	goutils.MustBeNil(err)

	mailbox.SendEmail(payload)
}
