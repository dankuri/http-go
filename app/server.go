package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		slog.Error("failed to bind to port 4221")
		os.Exit(1)
	}
	slog.Info("started listening on port 4221")

	var connID int

	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error("failed to accept conn", "err", err)
			os.Exit(1)
		}
		connID++
		go func() {
			err := handleConn(conn, connID)
			if err != nil {
				slog.Error("failed to handle conn", "err", err)
			}
		}()
	}
}

func handleConn(conn net.Conn, connID int) error {
	rd := bufio.NewReader(conn)
	req, err := ParseRequest(rd)
	if err != nil {
		return err
	}

	if req.Path == "/" {
		_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		slog.Info("responded 200", "connID", connID)
		return err
	} else if strings.HasPrefix(req.Path, "/echo") {
		resp, err := handleEcho(req)
		if err != nil {
			return fmt.Errorf("failed to handle echo req: %w", err)
		}
		return resp.Encode(conn)
	} else if req.Path == "/user-agent" {
		resp, err := handleUserAgent(req)
		if err != nil {
			return fmt.Errorf("failed to handle user-agent req: %w", err)
		}
		return resp.Encode(conn)
	}

	_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	slog.Info("responded 404", "connID", connID)

	return err
}

func handleEcho(req *HTTPRequest) (*HTTPResponse, error) {
	data, found := strings.CutPrefix(req.Path, "/echo/")
	if !found || len(data) == 0 {
		errMsg := "empty path"
		errResp := &HTTPResponse{
			Proto:      req.Proto,
			Status:     400,
			StatusText: "Bad Request",
			Headers: HTTPHeaders{
				"Content-Type":   "text/plain",
				"Content-Length": strconv.Itoa(len(errMsg)),
			},
			ResponseBody: []byte(errMsg),
		}
		return errResp, nil
	}

	resp := &HTTPResponse{
		Proto:      req.Proto,
		Status:     200,
		StatusText: "OK",
		Headers: HTTPHeaders{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(data)),
		},
		ResponseBody: []byte(data),
	}

	return resp, nil
}

func handleUserAgent(req *HTTPRequest) (*HTTPResponse, error) {
	userAgent, found := req.Headers["User-Agent"]
	if !found || len(userAgent) == 0 {
		errMsg := "empty User-Agent"
		errResp := &HTTPResponse{
			Proto:      req.Proto,
			Status:     400,
			StatusText: "Bad Request",
			Headers: HTTPHeaders{
				"Content-Type":   "text/plain",
				"Content-Length": strconv.Itoa(len(errMsg)),
			},
			ResponseBody: []byte(errMsg),
		}
		return errResp, nil
	}

	resp := &HTTPResponse{
		Proto:      req.Proto,
		Status:     200,
		StatusText: "OK",
		Headers: HTTPHeaders{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(userAgent)),
		},
		ResponseBody: []byte(userAgent),
	}

	return resp, nil
}
