package main

import (
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
	conn, err := l.Accept()
	if err != nil {
		slog.Error("failed to accept conn", "err", err)
		os.Exit(1)
	}
	_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if err != nil {
		slog.Error("failed to writing to conn", "err", err)
		os.Exit(1)
	}
}
