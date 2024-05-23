package main

import (
	"bufio"
	"fmt"
	"io"
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

	req := &HTTPRequest{
		Method: method,
		Path:   path,
		Proto:  proto,
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
