package main

import (
	_ "embed"

	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/google/mail/internal/files"
	"github.com/amp-labs/connectors/test/google/mail/internal/mailbox"
)

var (
	//go:embed email.json
	email []byte
)

func main() {
	flags := mailbox.ParseFlags("to", "from", "cc", "bcc")

	text := `This email has some text and 2 image attachments.
Also here are some emojis: 😁🥳🚴

Additionally the cc, bcc are included in this email.`

	html := `<div dir="ltr"><div dir="ltr"><div dir="ltr">This email has some text and 2 image attachments.<div>Also here are some emojis: 😁🥳🚴</div><div><br></div><div>Additionally the cc, bcc are included in this email.</div></div></div></div>`

	payload, err := mailbox.CreatePayload(email,
		mailbox.NewVar("To", flags.To),
		mailbox.NewVar("From", flags.From),
		mailbox.NewVar("Cc", flags.Cc),
		mailbox.NewVar("Bcc", flags.Bcc),
		mailbox.NewVar("Subject", "{002} Email with Image"),
		mailbox.NewEncodedVar("TEXT", []byte(text)),
		mailbox.NewEncodedVar("HTML", []byte(html)),
		mailbox.NewEncodedVar("Bridge", files.ImageBridge),
		mailbox.NewEncodedVar("Norway", files.ImageNorway),
	)
	goutils.MustBeNil(err)

	mailbox.SendEmail(payload)
}
