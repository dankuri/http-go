package main

import (
	"bufio"
	"fmt"
	"strings"
)

type HTTPMethod = string

const (
	GET  HTTPMethod = "GET"
	POST HTTPMethod = "POST"
)

type HTTPRequest struct {
	Method      HTTPMethod
	Path        string
	Proto       string
	Headers     map[string]string
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
