package main

import (
	"bufio"
	"log/slog"
	"net"
	"os"
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
	}

	_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	slog.Info("responded 404", "connID", connID)

	return err
}
