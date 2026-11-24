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

	text := `Hi,

This is a reply to a thread. It includes an embedded image as well as an
attached pdf and image.
[image: house.png]


Best Regard,
Test Development

On Mon, Nov 16, 2026 at 9:07 PM John Doe <
john.doe@test.com> wrote:

> Message comes from connector.Write(ObjectName: AMPERSAND-messages)
`

	html := `<div dir="ltr"><div>Hi,</div><div><br></div>This is a reply to a thread. It includes an embedded image as well as an attached pdf and image.<div><img src="cid:ii_mwjn008h0" alt="house.png" width="154" height="81"><br></div><div><br><div><br></div><div>Best Regard,</div><div>Test Development</div></div></div><br><div class="gmail_quote gmail_quote_container"><div dir="ltr" class="gmail_attr">On Mon, Nov 16, 2026 at 9:07 PM John Doe &lt;<a href="mailto:john@test.com">john@test.com</a>&gt; wrote:<br></div><blockquote class="gmail_quote" style="margin:0px 0px 0px 0.8ex;border-left:1px solid rgb(204,204,204);padding-left:1ex">Message comes from connector.Write(ObjectName: AMPERSAND-messages)</blockquote></div>`

	payload, err := mailbox.CreatePayload(email,
		mailbox.NewVar("To", flags.To),
		mailbox.NewVar("From", flags.From),
		mailbox.NewVar("Subject", "Re: {006} Responding to some email"),
		mailbox.NewVar("Reply", `<CAEJ5D9=w4y1EtSpkBo0b2yOLgZVsrG_kirT4d_Pi6RnLYGbYEw@mail.gmail.com>`),
		mailbox.NewEncodedVar("TEXT", []byte(text)),
		mailbox.NewEncodedVar("HTML", []byte(html)),
		mailbox.NewEncodedVar("House", files.ImageHouse),
		mailbox.NewEncodedVar("Norway", files.ImageNorway),
		mailbox.NewEncodedVar("Document", files.DocumentAboutGoogleDev),
	)
	goutils.MustBeNil(err)

	mailbox.SendEmail(payload)
}
