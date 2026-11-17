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
	payload, err := mailbox.CreatePayload(email,
		mailbox.NewVar("House", files.ImageHouse),
	)
	goutils.MustBeNil(err)

	mailbox.SendEmail(payload)
}
