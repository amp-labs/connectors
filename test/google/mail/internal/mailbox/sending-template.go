package mailbox

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"text/template" // nosemgrep: go.lang.security.audit.xss.import-text-template.import-text-template

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
)

func CreateDraft(payload map[string]any) {
	createEmail("AMPERSAND-draft", payload)
}

func SendEmail(payload map[string]any) {
	createEmail("AMPERSAND-messages", payload)
}

func createEmail(objectName string, payload map[string]any) {
	ctx := context.Background()
	conn := connTest.GetGoogleMailConnector(ctx)
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}

// NewVar creates a variable from a plain string without encoding.
// Use this for simple string values like email addresses (To, From, Cc, Bcc).
func NewVar(name string, value string) Variable {
	return Variable{
		key:   name,
		value: value,
	}
}

// NewEncodedVar creates a variable for a media with encoding.
// Use this for inlined or attached files.
func NewEncodedVar(name string, data []byte) Variable {
	value := base64.RawURLEncoding.EncodeToString(data)

	return NewVar(name, value)
}

type Variable struct {
	key   string
	value string
}

func CreatePayload(emailTemplate []byte, vars ...Variable) (map[string]any, error) {
	templateData := createTemplateData(vars)

	tmpl := template.Must(
		template.New("input").Parse(string(emailTemplate)),
	)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func createTemplateData(vars []Variable) map[string]any {
	result := make(map[string]any)
	for _, variable := range vars {
		result[variable.key] = variable.value
	}

	return result
}
