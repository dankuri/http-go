package main

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
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
	Method      HTTPMethod
	Path        string
	Proto       string
	Headers     HTTPHeaders
	RequestBody []byte
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

	req := &HTTPRequest{
		Method:  method,
		Path:    path,
		Proto:   proto,
		Headers: headers,
	}

	return req, nil
}

type HTTPResponse struct {
	Proto        string
	Status       uint16
	StatusText   string
	Headers      HTTPHeaders
	ResponseBody []byte
}

func (resp *HTTPResponse) Encode(w io.Writer) error {
	_, err := fmt.Fprintf(
		w,
		"%s %d %s\r\n%s\r\n%s",
		resp.Proto,
		resp.Status,
		resp.StatusText,
		resp.Headers,
		resp.ResponseBody,
	)

	return err
}
