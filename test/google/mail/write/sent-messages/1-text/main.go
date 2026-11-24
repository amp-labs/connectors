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

	payload, err := mailbox.CreatePayload(email,
		mailbox.NewVar("To", flags.To),
		mailbox.NewVar("From", flags.From),
		mailbox.NewVar("Subject", "{001} Plain and simple"),
		mailbox.NewEncodedVar("TEXT", []byte(`Simple text email sent via connector.`)),
	)
	goutils.MustBeNil(err)

	mailbox.SendEmail(payload)
}
