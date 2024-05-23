package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Dir string
}

func main() {
	dir := flag.String("directory", "./", "root directory for file server")
	flag.Parse()

	cfg := &Config{
		Dir: *dir,
	}

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
			err := handleConn(cfg, conn, connID)
			if err != nil {
				slog.Error("failed to handle conn", "err", err)
			}
		}()
	}
}

func handleConn(cfg *Config, conn net.Conn, connID int) error {
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
	} else if req.Method == GET && strings.HasPrefix(req.Path, "/files/") {
		resp, err := handleGetFile(cfg.Dir, req)
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

func handleGetFile(rootDir string, req *HTTPRequest) (*HTTPResponse, error) {
	filePath, _ := strings.CutPrefix(req.Path, "/files/")
	data, err := os.ReadFile(filepath.Join(rootDir, filePath))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Error("something is wrong with reading file", "err", err)
		}
		return NotFoundResp(), nil
	}

	resp := &HTTPResponse{
		Proto:      req.Proto,
		Status:     200,
		StatusText: "OK",
		Headers: HTTPHeaders{
			"Content-Type":   "application/octet-stream",
			"Content-Length": strconv.Itoa(len(data)),
		},
		ResponseBody: data,
	}

	return resp, nil
}
