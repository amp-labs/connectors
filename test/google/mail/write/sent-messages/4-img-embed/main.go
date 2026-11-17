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

This is MIME test 4.

The image below is embedded in the message body rather than attached
as a separate downloadable file.
[image: image.png]

After sending this message, inspect its raw source. Look for:

- multipart/related
- text/html
- image/png or image/jpeg
- Content-ID
- cid: reference in the HTML
- Content-Disposition: inline

The HTML should contain something conceptually like:

<img src="cid:some-image-id">

Regards,
MIME Explorer
`

	embedIdentifier := "ii_mwfmg0ju0"

	html := `<div dir="ltr"><div dir="ltr"><pre>Hello,

This is MIME test 4.

The image below is embedded in the message body rather than attached
as a separate downloadable file.

<img src="cid:` + embedIdentifier + `" alt="image.png" width="158" height="85">

After sending this message, inspect its raw source. Look for:

- multipart/related
- text/html
- image/png or image/jpeg
- Content-ID
- cid: reference in the HTML
- Content-Disposition: inline

The HTML should contain something conceptually like:

&lt;img src=&quot;cid:some-image-id&quot;&gt;

Regards,
MIME Explorer</pre></div>
</div>
`

	payload, err := mailbox.CreatePayload(email,
		mailbox.NewVar("To", flags.To),
		mailbox.NewVar("From", flags.From),
		mailbox.NewVar("Subject", "{004} Embedded image"),
		mailbox.NewVar("EmbedId", embedIdentifier),
		mailbox.NewEncodedVar("TEXT", []byte(text)),
		mailbox.NewEncodedVar("HTML", []byte(html)),
		mailbox.NewEncodedVar("House", files.ImageHouse),
	)
	goutils.MustBeNil(err)

	mailbox.SendEmail(payload)
}
