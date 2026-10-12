package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template" // nosemgrep: go.lang.security.audit.xss.import-text-template.import-text-template

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
)

// Blank lines are very important and are used by Gmail API.
var payloadTemplate = `From: {{.SenderEmail}}
To: {{.RecipientEmail}}
Subject: Hello

This is a test message.
`

func main() {
	from := flag.String("from", "", "sender email")
	to := flag.String("to", "", "recipient email")
	flag.Parse()

	fmt.Println(*from)
	fmt.Println(*to)

	if *from == "" || *to == "" {
		flag.Usage()
		os.Exit(1)
	}

	tmpl, err := template.New("json").Parse(payloadTemplate)
	if err != nil {
		utils.Fail("parsing template", "error", err)
	}

	tmplVars := map[string]string{
		"SenderEmail":    *from,
		"RecipientEmail": *to,
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, tmplVars); err != nil {
		utils.Fail("error creating payload from template", "error", err)
	}

	mimeText := strings.ReplaceAll(buf.String(), "\n", "\r\n")
	encodedMessage := base64.RawURLEncoding.EncodeToString([]byte(mimeText))

	payload := map[string]any{
		"raw": encodedMessage,
	}

	ctx := context.Background()
	conn := connTest.GetGoogleMailConnector(ctx)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "AMPERSAND-messages",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

}
