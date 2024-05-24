package main

import "strconv"

func OKResp(dataType string, body []byte) *HTTPResponse {
	return &HTTPResponse{
		Proto:      "HTTP/1.1",
		Status:     200,
		StatusText: "OK",
		Headers: HTTPHeaders{
			"Content-Type":   dataType,
			"Content-Length": strconv.Itoa(len(body)),
		},
		Body: body,
	}
}

func BadResp(msg string) *HTTPResponse {
	return &HTTPResponse{
		Proto:      "HTTP/1.1",
		Status:     400,
		StatusText: "Bad Request",
		Headers: HTTPHeaders{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(msg)),
		},
		Body: []byte(msg),
	}
}

func NotFoundResp() *HTTPResponse {
	return &HTTPResponse{
		Proto:      "HTTP/1.1",
		Status:     404,
		StatusText: "Not Found",
	}
}

func InternalErrResp(msg string) *HTTPResponse {
	return &HTTPResponse{
		Proto:      "HTTP/1.1",
		Status:     500,
		StatusText: "Internal Server Error",
		Headers: HTTPHeaders{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(msg)),
		},
		Body: []byte(msg),
	}
}
