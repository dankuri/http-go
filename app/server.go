package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"path/filepath"
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
			slog.Debug("finished serving connection", "connID", connID)
		}()
	}
}

func handleConn(cfg *Config, conn net.Conn, connID int) error {
	rd := bufio.NewReader(conn)
	for {
		req, err := ParseRequest(rd)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		slog.Debug("new request",
			"connID", connID,
			"method", req.Method,
			"path", req.Path,
		)

		var resp *HTTPResponse
		handler := "unknown"

		switch {
		case req.Path == "/":
			resp = &HTTPResponse{
				Proto:      "HTTP/1.1",
				Status:     200,
				StatusText: "OK",
			}
		case strings.HasPrefix(req.Path, "/echo"):
			handler = "echo"
			resp, err = handleEcho(req)
		case req.Path == "/user-agent":
			handler = "user-agent"
			resp, err = handleUserAgent(req)
		case req.Method == GET && strings.HasPrefix(req.Path, "/files/"):
			handler = "get files"
			resp, err = handleGetFile(cfg.Dir, req)
		case req.Method == POST && strings.HasPrefix(req.Path, "/files/"):
			handler = "post files"
			resp, err = handlePostFile(cfg.Dir, req)
		default:
			resp = &HTTPResponse{
				Proto:      "HTTP/1.1",
				Status:     404,
				StatusText: "Not Found",
			}
		}

		if err != nil {
			return fmt.Errorf("failed to handle %s req: %w", handler, err)
		}

		err = resp.Encode(conn)
		if err != nil {
			return err
		}
	}
}

func handleEcho(req *HTTPRequest) (*HTTPResponse, error) {
	data, found := strings.CutPrefix(req.Path, "/echo/")
	if !found || len(data) == 0 {
		return BadResp("empty path"), nil
	}

	return OKResp("text/plain", []byte(data)).MatchEncoding(req), nil
}

func handleUserAgent(req *HTTPRequest) (*HTTPResponse, error) {
	userAgent, found := req.Headers["User-Agent"]
	if !found || len(userAgent) == 0 {
		return BadResp("empty User-Agent"), nil
	}

	return OKResp("text/plain", []byte(userAgent)), nil
}

func handleGetFile(rootDir string, req *HTTPRequest) (*HTTPResponse, error) {
	fileName, _ := strings.CutPrefix(req.Path, "/files/")
	data, err := os.ReadFile(filepath.Join(rootDir, fileName))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			slog.Error("something is wrong with reading file", "err", err)
		}
		return NotFoundResp(), nil
	}

	return OKResp("application/octet-stream", data).MatchEncoding(req), nil
}

func handlePostFile(rootDir string, req *HTTPRequest) (*HTTPResponse, error) {
	fileName, _ := strings.CutPrefix(req.Path, "/files/")
	fullPath := filepath.Join(rootDir, fileName)
	file, err := os.Create(fullPath)
	if err != nil {
		return InternalErrResp("failed to create file"), nil
	}
	defer file.Close()

	_, err = file.Write(req.Body)
	if err != nil {
		return InternalErrResp("failed to save file"), nil
	}

	resp := &HTTPResponse{
		Proto:      req.Proto,
		Status:     201,
		StatusText: "Created",
	}

	return resp, nil
}
