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
	payload, err := mailbox.CreatePayload(email,
		mailbox.NewVar("Bridge", files.ImageBridge),
		mailbox.NewVar("Norway", files.ImageNorway),
	)
	goutils.MustBeNil(err)

	mailbox.SendEmail(payload)
}
