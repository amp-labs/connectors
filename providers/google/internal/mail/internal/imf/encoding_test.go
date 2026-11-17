package imf

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name              string
		jsonInput         string
		expectContains    []string
		expectBlocks      []string
		expectNotContains []string
	}{
		{
			name: "plain text email",
			jsonInput: `{
				"payload": {
					"mimeType": "text/plain",
					"headers": [
						{"name": "From", "value": "Alice <alice@example.com>"},
						{"name": "To", "value": "Bob <bob@example.com>"},
						{"name": "Subject", "value": "Hello World"}
					],
					"body": {
						"data": "SGVsbG8sIHRoaXMgaXMgYSBwbGFpbiB0ZXh0IGVtYWlsIGJvZHku"
					}
				}
			}`,
			expectContains: []string{
				`From: "Alice" <alice@example.com>`,
				`"Bob" <bob@example.com>`,
				`Subject: Hello World`,
				`Hello, this is a plain text email body.`,
			},
		},
		{
			name: "email with inline images",
			jsonInput: `{
				"payload": {
					"mimeType": "multipart/mixed",
					"headers": [
						{"name": "From", "value": "Alice <alice@example.com>"},
						{"name": "To", "value": "Bob <bob@example.com>"},
						{"name": "Cc", "value": "Carol <carol@example.com>"},
						{"name": "Bcc", "value": "Dan <dan@example.com>"},
						{"name": "Subject", "value": "Email with Images"}
					],
					"parts": [
						{
							"mimeType": "multipart/alternative",
							"parts": [
								{
									"mimeType": "text/plain",
									"body": {
										"data": "VGhpcyBlbWFpbCBoYXMgdGV4dCBhbmQgaW5saW5lIGltYWdlcy4"
									}
								},
								{
									"mimeType": "text/html",
									"body": {
										"data": "PGh0bWw-PGJvZHk-PHA-VGhpcyBlbWFpbCBoYXMgdGV4dCBhbmQgPGltZyBzcmM9ImNpZDppbWFnZTEiIGFsdD0iaW1hZ2UxLnBuZyI-IGlubGluZSBpbWFnZXMuPC9wPjwvYm9keT48L2h0bWw-"
									}
								}
							]
						},
						{
							"mimeType": "image/png",
							"filename": "image1.png",
							"headers": [
								{"name": "Content-ID", "value": "<image1>"}
							],
							"body": {
								"data": "ZHVtbXktaW1hZ2UtZGF0YQ"
							}
						},
						{
							"mimeType": "image/png",
							"filename": "image2.png",
							"headers": [
								{"name": "Content-ID", "value": "<image2>"},
								{"name": "Content-Disposition", "value": "inline; filename=\"image2.png\""}
							],
							"body": {
								"data": "ZHVtbXktaW1hZ2UtZGF0YQ"
							}
						}
					]
				}
			}`,
			expectContains: []string{
				`From: "Alice" <alice@example.com>`,
				`To: "Bob" <bob@example.com>`,
				`Cc: "Carol" <carol@example.com>`,
				`Bcc: "Dan" <dan@example.com>`,
				`Subject: Email with Images`,
			},
			expectBlocks: []string{
				// Text.
				`Content-Transfer-Encoding: quoted-printable
				Content-Type: text/plain; charset=UTF-8
				
				This email has text and inline images.`,
				// HTML.
				`Content-Transfer-Encoding: quoted-printable
				Content-Type: text/html; charset=UTF-8
				
				<html><body><p>This email has text and <img src=3D"cid:image1" alt=3D"image=
				1.png"> inline images.</p></body></html>`,
				// Inline 1. This happens because "cid:image1" is mentioned within HTML.
				`Content-Disposition: inline; filename="image1.png"
				Content-Id: <image1>
				Content-Transfer-Encoding: base64
				Content-Type: image/png
				
				ZHVtbXktaW1hZ2UtZGF0YQ==`,
				// Inline 2. This happens because "Content-Disposition" is set to "inline",
				// even though the HTML doesn't mention the ID.
				`Content-Disposition: inline; filename="image2.png"
				Content-Id: <image2>
				Content-Transfer-Encoding: base64
				Content-Type: image/png
				
				ZHVtbXktaW1hZ2UtZGF0YQ==`,
			},
		},
		{
			name: "HTML multipart email",
			jsonInput: `{
				"payload": {
					"mimeType": "multipart/alternative",
					"headers": [
						{"name": "From", "value": "Alice <alice@example.com>"},
						{"name": "To", "value": "Bob <bob@example.com>"},
						{"name": "Subject", "value": "HTML Format Test"}
					],
					"parts": [
						{
							"mimeType": "text/plain",
							"body": {
								"data": "VGhpcyBpcyB0aGUgcGxhaW4gdGV4dCBhbHRlcm5hdGl2ZS4"
							}
						},
						{
							"mimeType": "text/html",
							"body": {
								"data": "PGh0bWw-PGJvZHk-PHA-SFRNTCB2ZXJzaW9uIHdpdGggPGI-Ym9sZDwvYj4gdGV4dC48L3A-PC9ib2R5PjwvaHRtbD4"
							}
						}
					]
				}
			}`,
			expectContains: []string{
				`To: "Bob" <bob@example.com>`,
				`From: "Alice" <alice@example.com>`,
				`Subject: HTML Format Test`,
			},
			expectBlocks: []string{
				// Text.
				`Content-Transfer-Encoding: quoted-printable
				Content-Type: text/plain; charset=UTF-8
				
				This is the plain text alternative.`,
				// HTML.
				`Content-Transfer-Encoding: quoted-printable
				Content-Type: text/html; charset=UTF-8
				
				<html><body><p>HTML version with <b>bold</b> text.</p></body></html>`,
			},
		},
		{
			name: "email with PDF attachment",
			jsonInput: `{
				"payload": {
					"mimeType": "multipart/mixed",
					"headers": [
						{"name": "From", "value": "Alice <alice@example.com>"},
						{"name": "To", "value": "Bob <bob@example.com>"},
						{"name": "Subject", "value": "Email with PDF"}
					],
					"parts": [
						{
							"mimeType": "multipart/alternative",
							"parts": [
								{
									"mimeType": "text/plain",
									"body": {
										"data": "SGVsbG8sIHRoaXMgbWVzc2FnZSBoYXMgYSBQREYgYXR0YWNoZWQu"
									}
								},
								{
									"mimeType": "text/html",
									"body": {
										"data": "PGh0bWw-PGJvZHk-PHA-SGVsbG8sIHRoaXMgbWVzc2FnZSBoYXMgYSA8YSBocmVmPSJkb2MucGRmIj5QREYgYXR0YWNoZWQ8L2E-LjwvcD48L2JvZHk-PC9odG1sPg"
									}
								}
							]
						},
						{
							"mimeType": "application/pdf",
							"filename": "document.pdf",
							"headers": [
								{"name": "Content-Disposition", "value": "attachment; filename=\"document.pdf\""}
							],
							"body": {
								"data": "ZHVtbXktcGRmLWRhdGE"
							}
						}
					]
				}
			}`,
			expectContains: []string{
				`From: "Alice" <alice@example.com>`,
				`To: "Bob" <bob@example.com>`,
				`Subject: Email with PDF`,
				`Hello, this message has a PDF attached.`,
			},
			expectBlocks: []string{
				// Text.
				`Content-Transfer-Encoding: quoted-printable
				Content-Type: text/plain; charset=UTF-8
				
				Hello, this message has a PDF attached.`,
				// HTML.
				`Content-Transfer-Encoding: quoted-printable
				Content-Type: text/html; charset=UTF-8
				
				<html><body><p>Hello, this message has a <a href=3D"doc.pdf">PDF attached</=
				a>.</p></body></html>`,
				// Attachment.
				`Content-Disposition: attachment;
					filename="document.pdf"
				Content-Id: <document.pdf>
				Content-Transfer-Encoding: base64
				Content-Type: application/pdf
				
				ZHVtbXktcGRmLWRhdGE=`,
			},
		},
		{
			name: "reply with inline image and attachments",
			jsonInput: `{
				"payload": {
					"mimeType": "multipart/mixed",
					"headers": [
						{"name": "From", "value": "Alice <alice@example.com>"},
						{"name": "To", "value": "Bob <bob@example.com>"},
						{"name": "Subject", "value": "Re: Previous subject"},
						{"name": "In-Reply-To", "value": "<original@example.com>"},
						{"name": "References", "value": "<original@example.com>"}
					],
					"parts": [
						{
							"mimeType": "multipart/related",
							"parts": [
								{
									"mimeType": "multipart/alternative",
									"parts": [
										{
											"mimeType": "text/plain",
											"body": {
												"data": "VGhpcyBpcyBhIHJlcGx5IHRvIHRoZSBvcmlnaW5hbCBtZXNzYWdlLg"
											}
										},
										{
											"mimeType": "text/html",
											"body": {
												"data": "PGh0bWw-PGJvZHk-PHA-UmVwbHkgd2l0aCA8aW1nIHNyYz0iY2lkOnRlc3QtaW1hZ2UtaWQiIGFsdD0idGVzdC5wbmciPjwvcD48L2JvZHk-PC9odG1sPg"
											}
										}
									]
								},
								{
									"mimeType": "image/png",
									"filename": "test.png",
									"headers": [
										{"name": "Content-ID", "value": "<test-image-id>"}
									],
									"body": {
										"data": "ZHVtbXktaW1hZ2UtZGF0YQ"
									}
								}
							]
						},
						{
							"mimeType": "application/pdf",
							"filename": "document.pdf",
							"headers": [
								{"name": "Content-Disposition", "value": "attachment; filename=\"document.pdf\""}
							],
							"body": {
								"data": "ZHVtbXktcGRmLWRhdGE"
							}
						}
					]
				}
			}`,
			expectContains: []string{
				`From: "Alice" <alice@example.com>`,
				`To: "Bob" <bob@example.com>`,
				`In-Reply-To: <original@example.com>`,
				`References: <original@example.com>`,
				`Subject: Re: Previous subject`,
			},
			expectBlocks: []string{
				// Text.
				`Content-Transfer-Encoding: quoted-printable
				Content-Type: text/plain; charset=UTF-8
				
				This is a reply to the original message.`,
				// HTML.
				`Content-Transfer-Encoding: quoted-printable
				Content-Type: text/html; charset=UTF-8
				
				<html><body><p>Reply with <img src=3D"cid:test-image-id" alt=3D"test.png"><=
				/p></body></html>`,
				// Inline IMG.
				`Content-Disposition: inline; filename="test.png"
				Content-Id: <test-image-id>
				Content-Transfer-Encoding: base64
				Content-Type: image/png
				
				ZHVtbXktaW1hZ2UtZGF0YQ==`,
				// Attachment PDF.
				`Content-Disposition: attachment;
					filename="document.pdf"
				Content-Id: <document.pdf>
				Content-Transfer-Encoding: base64
				Content-Type: application/pdf
				
				ZHVtbXktcGRmLWRhdGE=`,
			},
		},
		{
			name: "draft-wrapped message",
			jsonInput: `{
				"message": {
					"payload": {
						"mimeType": "text/plain",
						"headers": [
							{"name": "From", "value": "Draft Sender <draft@example.com>"},
							{"name": "To", "value": "Recipient <recipient@example.com>"},
							{"name": "Subject", "value": "Draft Subject"}
						],
						"body": {
							"data": "VGhpcyBpcyBhIGRyYWZ0IG1lc3NhZ2UgYm9keS4"
						}
					}
				}
			}`,
			expectContains: []string{
				`From: "Draft Sender" <draft@example.com>`,
				`To: "Recipient" <recipient@example.com>`,
				`Subject: Draft Subject`,
				`This is a draft message body.`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse JSON string into map[string]any, which is the input of FormatMessage.
			var msg map[string]any
			if err := json.Unmarshal([]byte(tt.jsonInput), &msg); err != nil {
				t.Fatalf("failed to unmarshal JSON input: %v", err)
			}

			// Call FormatMessage.
			result, isEncoded, err := FormatMessage(msg)
			if err != nil {
				t.Fatalf("FormatMessage returned error: %v", err)
			}

			if isEncoded {
				t.Fatal("FormatMessage should not return encoded text when no 'raw' field is present")
			}

			resultStr := string(result)

			// Check that expected substrings are present.
			for _, expected := range tt.expectContains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("expected output to contain %q, but it was not found", expected)
				}
			}

			// Check that expected contiguous blocks are present.
			for _, block := range tt.expectBlocks {
				if !containsBlock(resultStr, block) {
					t.Errorf("expected output to contain block:\n%v\nbut it was not found", block)
				}
			}

			// Check that excluded substrings are absent.
			for _, excluded := range tt.expectNotContains {
				if strings.Contains(resultStr, excluded) {
					t.Errorf("expected output to NOT contain %q, but it was found", excluded)
				}
			}
		})
	}
}

// containsBlock checks that all lines in the expected block appear as a
// contiguous sequence in the result. Tabs and leading/trailing whitespace
// on each line are trimmed when comparing, so the test is robust against
// formatting differences in the output. Empty lines in the expected block
// are preserved to match the MIME structure (e.g., blank line between headers
// and body).
func containsBlock(result, expectedBlock string) bool {
	normalize := func(s string) string {
		// Trim leading and trailing whitespace (including tabs) from each line.
		return strings.TrimSpace(s)
	}
	expectedLines := strings.Split(expectedBlock, "\n")
	resultLines := strings.Split(result, "\n")

	var expected []string
	for _, line := range expectedLines {
		expected = append(expected, normalize(line))
	}
	if len(expected) == 0 {
		return true
	}

	var resultNorm []string
	for _, line := range resultLines {
		resultNorm = append(resultNorm, normalize(line))
	}

	for i := 0; i <= len(resultNorm)-len(expected); i++ {
		match := true
		for j, exp := range expected {
			if exp == "" {
				// Empty line in expected block matches any empty line in result.
				if resultNorm[i+j] != "" {
					match = false
					break
				}
			} else {
				// Non-empty line: expected must be a prefix of the result line.
				matched, err := regexp.MatchString("^"+regexp.QuoteMeta(exp)+".*", resultNorm[i+j])
				if err != nil || !matched {
					match = false
					break
				}
			}
		}
		if match {
			return true
		}
	}
	return false
}
