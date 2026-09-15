package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/template" // nosemgrep: go.lang.security.audit.xss.import-text-template.import-text-template

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/microsoft"
	"github.com/amp-labs/connectors/test/utils"
)

var payloadTemplate = `
{
    "message": {
        "subject": "This message will be created and sent right away via connector Write.",
        "body": {
            "contentType": "html",
            "content": "<html><head>\r\n<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\"><style type=\"text/css\" style=\"display:none\">\r\n<!--\r\np\r\n\t{margin-top:0;\r\n\tmargin-bottom:0}\r\n-->\r\n</style></head><body dir=\"ltr\"><div style=\"font-family:Aptos,Aptos_EmbeddedFont,Aptos_MSFontService,Calibri,Helvetica,sans-serif; font-size:12pt; color:rgb(0,0,0)\"><br></div><div class=\"elementToProof\" style=\"font-family:Aptos,Aptos_EmbeddedFont,Aptos_MSFontService,Calibri,Helvetica,sans-serif; font-size:12pt; color:rgb(0,0,0)\">Blah blah blah.</div><div class=\"elementToProof\" style=\"font-family:Aptos,Aptos_EmbeddedFont,Aptos_MSFontService,Calibri,Helvetica,sans-serif; font-size:12pt; color:rgb(0,0,0)\"><br></div></body></html>"
        },
        "sender": {
            "emailAddress": {
                "name": "{{.Name}}",
                "address": "{{.SenderEmail}}"
            }
        },
        "from": {
            "emailAddress": {
                "name": "{{.Name}}",
                "address": "{{.SenderEmail}}"
            }
        },
        "toRecipients": [
            {
                "emailAddress": {
                    "name": "{{.RecipientEmail}}",
                    "address": "{{.RecipientEmail}}"
                }
            }
        ]
    }
}
`

func main() {
	name := flag.String("name", "", "sender name")
	from := flag.String("from", "", "sender email")
	to := flag.String("to", "", "recipient email")
	flag.Parse()

	fmt.Println(*name)
	fmt.Println(*from)
	fmt.Println(*to)

	if *name == "" || *from == "" || *to == "" {
		flag.Usage()
		os.Exit(1)
	}

	tmpl, err := template.New("json").Parse(payloadTemplate)
	if err != nil {
		utils.Fail("parsing template", "error", err)
	}

	tmplVars := map[string]string{
		"Name":           *name,
		"SenderEmail":    *from,
		"RecipientEmail": *to,
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, tmplVars); err != nil {
		utils.Fail("error creating payload from template", "error", err)
	}

	var payload map[string]any
	if err = json.Unmarshal(buf.Bytes(), &payload); err != nil {
		utils.Fail("error converting payload to JSON object", "error", err)
	}

	ctx := context.Background()
	conn := connTest.GetMicrosoftGraphConnector(ctx)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "AMPERSAND-sentMessages",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)

}
