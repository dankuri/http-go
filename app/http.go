package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/textproto"
	"slices"
	"strconv"
	"strings"
)

type HTTPMethod = string

const (
	GET  HTTPMethod = "GET"
	POST HTTPMethod = "POST"
)

type HTTPHeaders map[string]string

func (h HTTPHeaders) String() string {
	builder := new(strings.Builder)
	for header, value := range h {
		_, err := fmt.Fprintf(builder, "%s: %s\r\n", header, value)
		if err != nil {
			return ""
		}
	}

	return builder.String()
}

type HTTPRequest struct {
	Method  HTTPMethod
	Path    string
	Proto   string
	Headers HTTPHeaders
	Body    []byte
}

func ParseRequest(r *bufio.Reader) (*HTTPRequest, error) {
	reqLine, err := r.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to parse request line: %w", err)
	}
	reqLine, found := strings.CutSuffix(reqLine, "\r\n")
	if !found {
		return nil, ErrInvalidFormat
	}

	reqLineSplit := strings.Split(reqLine, " ")
	method := reqLineSplit[0]
	path := reqLineSplit[1]
	proto := reqLineSplit[2]

	switch method {
	case GET, POST:
	default:
		return nil, fmt.Errorf("unknown method")
	}

	headers := make(HTTPHeaders)

	for {
		headerRaw, err := r.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to parse headers: %w", err)
		}
		headerRaw, found = strings.CutSuffix(headerRaw, "\r\n")
		if !found {
			return nil, ErrInvalidFormat
		}
		if headerRaw == "" {
			break
		}
		parts := strings.SplitN(headerRaw, ": ", 2)
		if len(parts) != 2 {
			return nil, ErrInvalidFormat
		}
		header := textproto.CanonicalMIMEHeaderKey(parts[0])
		headers[header] = parts[1]
	}

	var body []byte

	lenStr, found := headers["Content-Length"]
	if found {
		length, err := strconv.Atoi(lenStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Content-Lenght: %w", err)
		}

		body = make([]byte, length)
		n, err := r.Read(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		if n != length {
			return nil, fmt.Errorf("wrong Content-Lenght")
		}
	}

	req := &HTTPRequest{
		Method:  method,
		Path:    path,
		Proto:   proto,
		Headers: headers,
		Body:    body,
	}

	return req, nil
}

type HTTPResponse struct {
	Proto      string
	Status     uint16
	StatusText string
	Headers    HTTPHeaders
	Body       []byte
}

func (resp *HTTPResponse) MatchEncoding(req *HTTPRequest) *HTTPResponse {
	acceptEncList := strings.Split(req.Headers["Accept-Encoding"], ", ")
	if !slices.Contains(acceptEncList, "gzip") {
		return resp
	}

	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	_, err := w.Write(resp.Body)
	if err != nil {
		slog.Error("failed to compress body (write)", "err", err)
		return resp
	}
	err = w.Close()
	if err != nil {
		slog.Error("failed to compress body (close)", "err", err)
		return resp
	}

	resp.Body = buf.Bytes()
	resp.Headers["Content-Encoding"] = "gzip"
	resp.Headers["Content-Length"] = strconv.Itoa(len(resp.Body))

	return resp
}

func (resp *HTTPResponse) Encode(w io.Writer) error {
	_, err := fmt.Fprintf(
		w,
		"%s %d %s\r\n%s\r\n%s",
		resp.Proto,
		resp.Status,
		resp.StatusText,
		resp.Headers,
		resp.Body,
	)

	return err
}
