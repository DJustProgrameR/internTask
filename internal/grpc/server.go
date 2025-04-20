// Package grpc это grpc сервер
package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"internshipPVZ/internal/grpc/handler"
	pb "internshipPVZ/internal/grpc/models"
	"log"
	"net"
)

// Server -
type Server struct {
	lis net.Listener
	s   *grpc.Server
}

// NewServer конструктор
func NewServer(uc handler.GetPvzUseCase) *Server {
	if uc == nil {
		log.Fatalf("NewServer initialization failed: GetPvzUseCase is nil")
	}
	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterPVZServiceServer(s, handler.NewPVZServiceServer(uc))
	return &Server{s: s}
}

// Listen grpc слушает на порту
func (s *Server) Listen(portString string) error {
	lis, err := net.Listen("tcp", portString)
	if err != nil {
		return err
	}
	s.lis = lis
	if err = s.s.Serve(lis); err != nil {
		return err
	}
	return nil
}

// Shutdown выключение grpc
func (s *Server) Shutdown() error {
	return s.lis.Close()
}
