package mailbox

import (
	"flag"
	"log"
)

type EmailFlags struct {
	To   string
	From string
	Cc   string
	Bcc  string
}

func ParseFlags(required ...string) EmailFlags {
	to := flag.String("to", "", "Recipient email address")
	from := flag.String("from", "", "Sender email address")
	cc := flag.String("cc", "", "CC email address")
	bcc := flag.String("bcc", "", "BCC email address")

	flag.Parse()

	result := EmailFlags{
		To:   *to,
		From: *from,
		Cc:   *cc,
		Bcc:  *bcc,
	}

	values := map[string]string{
		"to":   result.To,
		"from": result.From,
		"cc":   result.Cc,
		"bcc":  result.Bcc,
	}

	for _, name := range required {
		if values[name] == "" {
			log.Fatalf("flag -%s is required", name)
		}
	}

	return result
}
