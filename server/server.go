package server

import (
	"context"
	"fmt"
	"github.com/Streamlet/NoteIsSite/global"
	"net"
	"net/http"
)

type HttpServer interface {
	Serve() error
	Shutdown() error
}

func NewPortServer(port uint, noteDir string, templateDir string) (HttpServer, error) {
	var err error
	server := new(portServer)
	server.Handler, err = newRouter(noteDir, templateDir)
	server.port = port
	return server, err
}

func NewSockServer(sock string, noteDir string, templateDir string) (HttpServer, error) {
	var err error
	server := new(sockServer)
	server.Handler, err = newRouter(noteDir, templateDir)
	server.sock = sock
	return server, err
}

type portServer struct {
	http.Server
	port uint
}

type sockServer struct {
	http.Server
	sock string
}

func serve(s *http.Server, l net.Listener) {
	go func() {
		err := s.Serve(l)
		if err != http.ErrServerClosed {
			global.GetErrorChan() <- err
		}
	}()
}

func shutdown(s *http.Server) error {
	return s.Shutdown(context.Background())
}

func (s portServer) Serve() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	serve(&s.Server, l)
	return nil
}

func (s portServer) Shutdown() error {
	return shutdown(&s.Server)
}

func (s sockServer) Serve() error {
	l, err := net.Listen("unix", s.sock)
	if err != nil {
		return err
	}
	serve(&s.Server, l)
	return nil
}

func (s sockServer) Shutdown() error {
	return shutdown(&s.Server)
}
