package main

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/google/mail/internal/mailbox"
)

//go:embed email.json
var email []byte

func main() {
	flags := mailbox.ParseFlags("to", "from")

	text := `MIME Internet Message Format test

This message is intended to become an HTML MIME part.

Hello,

This is a *rich-text HTML email* created to inspect how Gmail represents
formatting in the resulting MIME message.

It includes several HTML features that may produce a multipart/alternative
structure, with an HTML representation and possibly a plain-text fallback.

   - Bold and italic text
   - A hyperlink
   - A bullet list
   - A simple table
   - Emoji: âœ… ðŸ“§ ðŸ§ª

Visit the RFC 5322 Internet Message Format specification
<https://datatracker.ietf.org/doc/html/rfc5322>for the underlying
email-header format.
MIME concept What to look for
HTML body text/html
Plain-text fallback text/plain
Alternative parts multipart/alternative

Regards,
*MIME Explorer*
`

	html := `<div dir="ltr"><div dir="ltr"><div dir="ltr"><table role="presentation" width="600" cellpadding="0" cellspacing="0" style="font-size:medium;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial;border-radius:8px"><tbody><tr><td style="padding:24px;border-bottom:1px solid rgb(229,231,235)"><h1 style="margin:0px;font-size:22px;color:rgb(17,17,17)">MIME Internet Message Format test</h1><p style="margin:8px 0px 0px;font-size:14px;color:rgb(85,85,85)">This message is intended to become an HTML MIME part.</p></td></tr><tr><td style="padding:24px"><p style="margin:0px 0px 16px;font-size:15px;line-height:1.5">Hello,</p><p style="margin:0px 0px 16px;font-size:15px;line-height:1.5">This is a <strong>rich-text HTML email</strong> created to inspect how Gmail represents formatting in the resulting MIME message.</p><p style="margin:0px 0px 16px;font-size:15px;line-height:1.5">It includes several HTML features that may produce a <code style="background:rgb(241,243,244);padding:2px 4px;border-radius:3px">multipart/alternative</code> structure, with an HTML representation and possibly a plain-text fallback.</p><ul style="margin:0px 0px 16px;padding-left:20px;font-size:15px;line-height:1.6"><li>Bold and italic text</li><li>A hyperlink</li><li>A bullet list</li><li>A simple table</li><li>Emoji: âœ… ðŸ“§ ðŸ§ª</li></ul><p style="margin:0px 0px 16px;font-size:15px;line-height:1.5">Visit the <a href="https://datatracker.ietf.org/doc/html/rfc5322" style="color:rgb(26,115,232)" target="_blank">RFC 5322 Internet Message Format specification </a>for the underlying email-header format.</p><table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border-collapse:collapse;font-size:14px"><tbody><tr style="background:rgb(241,243,244)"><th align="left" style="padding:8px;border:1px solid rgb(218,220,224)">MIME concept</th><th align="left" style="padding:8px;border:1px solid rgb(218,220,224)">What to look for</th></tr><tr><td style="padding:8px;border:1px solid rgb(218,220,224)">HTML body</td><td style="padding:8px;border:1px solid rgb(218,220,224)"><code>text/html</code></td></tr><tr><td style="padding:8px;border:1px solid rgb(218,220,224)">Plain-text fallback</td><td style="padding:8px;border:1px solid rgb(218,220,224)"><code>text/plain</code></td></tr><tr><td style="padding:8px;border:1px solid rgb(218,220,224)">Alternative parts</td><td style="padding:8px;border:1px solid rgb(218,220,224)"><code>multipart/alternative</code></td></tr></tbody></table><p style="margin:24px 0px 0px;font-size:15px;line-height:1.5">Regards,<br><strong>MIME Explorer</strong></p></td></tr></tbody></table></div></div></div>`

	payload, err := mailbox.CreatePayload(email,
		mailbox.NewVar("To", flags.To),
		mailbox.NewVar("From", flags.From),
		mailbox.NewVar("Subject", "{003} HTML Format Test"),
		mailbox.NewEncodedVar("TEXT", []byte(text)),
		mailbox.NewEncodedVar("HTML", []byte(html)),
	)
	goutils.MustBeNil(err)

	mailbox.SendEmail(payload)
}
