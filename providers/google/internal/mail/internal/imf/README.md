# IMF

The `imf` package bridges Google's [Gmail API Message format](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages) and the Internet Message Format (IMF).

It reconstructs an RFC 2822-compliant MIME message from a Gmail message representation
including headers, text and HTML bodies, attachments, and embedded content.

The generated message can be supplied to the Gmail API using the `raw` field.

## Usage

Use `FormatMessage` to convert a Gmail message representation into IMF bytes:

```go
text, isEncoded, err := imf.FormatMessage(gmailMessage)
if err != nil {
    return err
}

if !isEncoded {
    text = base64.URLEncoding.EncodeToString(text)
}

msg := gmailMessagePayload{
    Raw: string(text),
}
```

If the input already contains a non-empty `raw` field, it is returned unchanged. Otherwise, the package reconstructs the MIME message from the supplied Gmail message representation.

## Message structure

The input follows the Gmail API `Message` resource structure. The relevant fields are:

- `payload.headers` — message headers such as `From`, `To`, `Subject`, and `Date`.
- `payload.parts` — MIME body parts and attachments.
- `payload.body.data` — Base64URL-encoded body content.
- `payload.parts[].filename` — attachment filename.
- `payload.parts[].mimeType` — MIME content type.

Body parts can be nested within multipart containers. The package discovers plain-text and HTML bodies and extracts attachments from the message structure.

## Attachments and embedded content

The distinction between an ordinary attachment and embedded content is important when constructing the input message.

**Regular attachments**

Provide a filename and the attachment data. The attachment is treated as a downloadable file.

**Embedded content**

To include an image or another resource in the HTML body, provide its `Content-ID` and reference it from the HTML using a `cid:` URL.

For example, the attachment part should include:

```json
{
  "filename": "image.png",
  "mimeType": "image/png",
  "headers": [
    {
      "name": "Content-ID",
      "value": "<image123>"
    }
  ],
  "body": {
    "data": "<base64url-encoded-image-data>"
  }
}
```

The HTML body should reference the same identifier:

```html
<img src="cid:image123" alt="Embedded image">
```

An attachment is also treated as inline when its `Content-Disposition` header explicitly specifies `inline`.

In other words, **either an explicit inline disposition or a matching `cid:` reference in the HTML body identifies embedded content**.

## Important notes

- Body and attachment data must be Base64URL-encoded as expected by the Gmail API. Supplied via `data` property.
- The package reconstructs the MIME message; it does not guarantee byte-for-byte reproduction of Gmail's original MIME serialization.
- Gmail-specific metadata such as `id`, `threadId`, `labelIds`, and `historyId` is not included in the generated IMF message.
- When the input contains `raw`, no reconstruction is performed.
