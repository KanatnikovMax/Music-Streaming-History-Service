package grpc

import (
	"MusicStreamingHistoryService/internal/service"
	pb "MusicStreamingHistoryService/pkg/proto/listening_history"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ListeningHistoryHandler struct {
	pb.UnimplementedListeningHistoryApiServer
	svc service.ListeningHistoryService
}

func NewListeningHistoryHandler(svc service.ListeningHistoryService) *ListeningHistoryHandler {
	return &ListeningHistoryHandler{svc: svc}
}

func (h *ListeningHistoryHandler) GetUserListeningHistory(
	ctx context.Context,
	req *pb.GetUserListeningHistoryRequest,
) (*pb.GetUserListeningHistoryResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id: %v", err)
	}

	items, err := h.svc.GetUserHistory(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user history: %v", err)
	}

	pbItems := make([]*pb.ListeningHistoryItem, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, &pb.ListeningHistoryItem{
			EventId:       item.EventID.String(),
			UserId:        item.UserID.String(),
			SongId:        item.SongID.String(),
			ListenedAtUtc: timestamppb.New(item.ListenedAtUtc),
		})
	}

	return &pb.GetUserListeningHistoryResponse{Items: pbItems}, nil
}
