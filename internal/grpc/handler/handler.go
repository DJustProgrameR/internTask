// Package handler это хэндлер grpc
package handler

import (
	"context"
	pb "internshipPVZ/internal/grpc/models"
	"time"
)

const (
	cancelContextTime time.Duration = time.Millisecond * 5000
)

// GetPvzUseCase --
type GetPvzUseCase interface {
	Get(ctx context.Context) *pb.GetPVZListResponse
}

// PVZServiceServer grpc сервис
type PVZServiceServer struct {
	pb.UnimplementedPVZServiceServer
	getPVZUseCase GetPvzUseCase
}

// NewPVZServiceServer конструктор
func NewPVZServiceServer(uc GetPvzUseCase) *PVZServiceServer {
	return &PVZServiceServer{getPVZUseCase: uc}
}

// GetPVZList возвращает лист пвз
func (s *PVZServiceServer) GetPVZList(ctx context.Context, _ *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	contWithTimeout, cancel := context.WithTimeout(ctx, cancelContextTime)
	defer cancel()

	resp := s.getPVZUseCase.Get(contWithTimeout)

	return resp, nil
}
