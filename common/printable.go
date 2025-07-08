package common

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

const truncationLength = 512 * 1024 // 512 KB

func logRequestWithBody(logger *slog.Logger, req *http.Request, method, id, fullURL string, body []byte) {
	headers := redactSensitiveRequestHeaders(GetRequestHeaders(req))

	logger = logger.With(
		"method", method,
		"url", fullURL,
		"correlationId", id,
		"headers", headers)

	payload, err := PrintableRequest(req, body)
	if err != nil {
		logger.Error("Error creating printable request", "error", err)

		return
	}

	truncatedBody, err := payload.Truncate(truncationLength)
	if err != nil {
		logger.Error("Error truncating request body", "error", err)

		return
	}

	logger.Debug("HTTP request", "body", truncatedBody)
}

func logRequestWithoutBody(logger *slog.Logger, req *http.Request, method, id, fullURL string) {
	headers := redactSensitiveRequestHeaders(GetRequestHeaders(req))

	logger = logger.With(
		"method", method,
		"url", fullURL,
		"correlationId", id,
		"headers", headers)

	logger.Debug("HTTP request")
}

func logResponseWithoutBody(logger *slog.Logger, res *http.Response, method, id, fullURL string) {
	headers := redactSensitiveResponseHeaders(GetResponseHeaders(res))

	logger = logger.With(
		"method", method,
		"url", fullURL,
		"correlationId", id,
		"headers", headers)

	logger.Debug("HTTP response")
}

func logResponseWithBody(logger *slog.Logger, res *http.Response, method, id, fullURL string, body []byte) {
	headers := redactSensitiveResponseHeaders(GetResponseHeaders(res))

	logger = logger.With(
		"method", method,
		"url", fullURL,
		"correlationId", id,
		"headers", headers)

	payload, err := PrintableResponse(res, body)
	if err != nil {
		logger.Error("Error creating printable response", "error", err)

		return
	}

	truncatedBody, err := payload.Truncate(truncationLength)
	if err != nil {
		logger.Error("Error truncating response body", "error", err)

		return
	}

	logger.Debug("HTTP response", "body", truncatedBody)
}

// PrintablePayload represents a payload that can be printed or displayed.
// It contains the content, its length, and whether it is base64 encoded.
// It also includes a truncated length for cases where the content is too large.
type PrintablePayload struct {
	Base64          bool   `json:"base64,omitempty"`
	Content         string `json:"content"`
	Length          int64  `json:"length"`
	TruncatedLength int64  `json:"truncatedLength,omitempty"`
}

func (p *PrintablePayload) String() string {
	if p == nil {
		return "<nil>"
	}

	if p.IsEmpty() {
		return ""
	}

	if p.IsBase64() {
		return "base64:" + p.Content
	}

	return p.Content
}

func (p *PrintablePayload) IsEmpty() bool {
	return p == nil || (p.Content == "" && p.Length == 0)
}

func (p *PrintablePayload) IsBase64() bool {
	return p != nil && p.Base64
}

func (p *PrintablePayload) GetContent() string {
	if p == nil {
		return ""
	}

	return p.Content
}

func (p *PrintablePayload) GetContentBytes() ([]byte, error) {
	if p == nil {
		return nil, nil //nolint:nilnil
	}

	if p.IsBase64() {
		return base64.StdEncoding.DecodeString(p.Content)
	}

	return []byte(p.Content), nil
}

func (p *PrintablePayload) GetLength() int64 {
	if p == nil {
		return 0
	}

	return p.Length
}

func (p *PrintablePayload) IsTruncated() bool {
	if p == nil {
		return false
	}

	return p.GetTruncatedLength() < p.Length
}

func (p *PrintablePayload) Clone() *PrintablePayload {
	if p == nil {
		return nil
	}

	return &PrintablePayload{
		Base64:          p.Base64,
		Content:         p.Content,
		Length:          p.Length,
		TruncatedLength: p.TruncatedLength,
	}
}

func (p *PrintablePayload) Truncate(size int64) (*PrintablePayload, error) {
	if p == nil || size < 0 {
		return nil, nil //nolint:nilnil
	}

	if size >= p.Length || size >= p.GetTruncatedLength() {
		// No truncation needed, just return the original
		return p, nil
	}

	cloned := p.Clone()

	if p.IsBase64() {
		bts, err := p.GetContentBytes()
		if err != nil {
			return nil, fmt.Errorf("error getting content bytes: %w", err)
		}

		cloned.TruncatedLength = size
		truncated := bts[:size]
		cloned.Content = base64.StdEncoding.EncodeToString(truncated)
	} else {
		cloned.Content = cloned.Content[:size]

		// String truncation vs byte truncation may disagree in length (due
		// to multibyte characters), so we need to ensure the length is correct.
		cloned.TruncatedLength = int64(len([]byte(cloned.Content)))
	}

	return cloned, nil
}

func (p *PrintablePayload) GetTruncatedLength() int64 {
	if p == nil {
		return 0
	}

	if p.TruncatedLength > 0 {
		return p.TruncatedLength
	}

	// If not set, use the full length
	return p.Length
}

// PrintableRequest creates a PrintablePayload from an HTTP request.
// The body parameter can be nil, in which case it will read the request body.
func PrintableRequest(req *http.Request, body []byte) (*PrintablePayload, error) {
	return getBodyAsPrintable(&requestContentReader{
		Request:   req,
		BodyBytes: body,
	})
}

// PrintableResponse creates a PrintablePayload from an HTTP response.
// The body parameter can be nil, in which case it will read the response body.
func PrintableResponse(resp *http.Response, body []byte) (*PrintablePayload, error) {
	return getBodyAsPrintable(&responseContentReader{
		Response:  resp,
		BodyBytes: body,
	})
}

type requestContentReader struct {
	Request   *http.Request
	BodyBytes []byte
}

func (r *requestContentReader) GetBody() io.ReadCloser {
	if r.Request == nil {
		return nil
	}

	if r.BodyBytes != nil {
		return io.NopCloser(bytes.NewReader(r.BodyBytes))
	}

	return r.Request.Body
}

func (r *requestContentReader) GetHeaders() http.Header {
	if r.Request == nil {
		return nil
	}

	return r.Request.Header
}

func (r *requestContentReader) SetBody(body io.ReadCloser) {
	if r.Request == nil {
		return
	}

	r.BodyBytes = nil // Clear cached bytes if we set a new body
	r.Request.Body = body
}

type responseContentReader struct {
	Response  *http.Response
	BodyBytes []byte
}

func (r *responseContentReader) GetBody() io.ReadCloser {
	if r.Response == nil {
		return nil
	}

	if r.BodyBytes != nil {
		return io.NopCloser(bytes.NewReader(r.BodyBytes))
	}

	return r.Response.Body
}

func (r *responseContentReader) GetHeaders() http.Header {
	if r.Response == nil {
		return nil
	}

	return r.Response.Header
}

func (r *responseContentReader) SetBody(body io.ReadCloser) {
	if r.Response == nil {
		return
	}

	r.Response.Body = body
	r.BodyBytes = nil
}

type bodyContentReader interface {
	GetBody() io.ReadCloser
	GetHeaders() http.Header
	SetBody(body io.ReadCloser)
}

func isPrintableMimeType(mimeType string) bool {
	// Check if the MIME type is text-based or a known printable format
	return strings.HasPrefix(mimeType, "text/") ||
		strings.HasSuffix(mimeType, "+json") ||
		strings.HasSuffix(mimeType, "+xml") ||
		mimeType == "application/json" ||
		mimeType == "application/xml" ||
		mimeType == "application/javascript" ||
		mimeType == "application/x-www-form-urlencoded"
}

func peekBody(bcr bodyContentReader) ([]byte, error) {
	if bcr == nil || bcr.GetBody() == nil {
		return nil, nil
	}

	body := bcr.GetBody()

	// Read the body without closing it
	var buf bytes.Buffer
	tee := io.TeeReader(body, &buf)
	data, err := io.ReadAll(tee)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Restore the body for further use
	bcr.SetBody(io.NopCloser(&buf))

	return data, nil
}

// getBodyAsPrintable checks if the HTTP response body is probably printable text.
func getBodyAsPrintable(br bodyContentReader) (*PrintablePayload, error) { //nolint:funlen
	if br == nil || br.GetBody() == nil {
		return nil, nil //nolint:nilnil
	}

	// Check MIME type
	contentType := br.GetHeaders().Get("Content-Type")

	mimeType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// If parsing fails, fallback to sniffing the content
		mimeType = ""
	}

	charsetStr := strings.ToLower(params["charset"])

	rawData, err := peekBody(br)
	if err != nil {
		return nil, fmt.Errorf("error peeking response body: %w", err)
	}

	if mimeType != "" && !isPrintableMimeType(mimeType) {
		return &PrintablePayload{
			Base64:  true,
			Content: base64.StdEncoding.EncodeToString(rawData),
			Length:  int64(len(rawData)),
		}, nil
	}

	// Decode to UTF-8
	decodedReader, err := charset.NewReaderLabel(charsetStr, bytes.NewReader(rawData))
	if err != nil {
		// If charset is unknown or not provided, fallback to UTF-8
		decodedReader = bytes.NewReader(rawData)
	}

	decodedData, err := io.ReadAll(transform.NewReader(decodedReader, nil))
	if err != nil {
		return nil, err
	}

	// Check UTF-8 validity (paranoia)
	if !utf8.Valid(decodedData) {
		return &PrintablePayload{
			Base64:  true,
			Content: base64.StdEncoding.EncodeToString(rawData),
			Length:  int64(len(rawData)),
		}, nil
	}

	// Check printability (sample max N bytes)
	const maxCheckLen = 1024

	checkLen := len(decodedData)

	if checkLen > maxCheckLen {
		checkLen = maxCheckLen
	}
	sample := decodedData[:checkLen]

	printable := 0
	total := 0
	for len(sample) > 0 {
		r, size := utf8.DecodeRune(sample)
		sample = sample[size:]
		total++
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			printable++
		}
	}

	if total == 0 {
		return &PrintablePayload{
			Content: "",
			Length:  0,
		}, nil
	}

	// Heuristic: 95%+ means printable
	isPrintable := float64(printable)/float64(total) > 0.95 //nolint:mnd

	if isPrintable {
		return &PrintablePayload{
			Content: string(decodedData),
			Length:  int64(len(decodedData)),
		}, nil
	}

	return &PrintablePayload{
		Base64:  true,
		Content: base64.StdEncoding.EncodeToString(rawData),
		Length:  int64(len(rawData)),
	}, nil
}
