package grpc

import (
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/transport"
	api "auth-service/pkg/api/auth/v1"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	api.UnimplementedAuthServiceServer
	as transport.AuthService
}

func NewAuthHandler(s transport.AuthService) *AuthHandler {
	return &AuthHandler{as: s}
}
func (h *AuthHandler) Login(ctx context.Context, request *api.LoginRequest) (*api.LoginResponse, error) {
	token, err := h.as.Login(ctx, request.GetEmail(), request.GetPassword())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "grpc.Login %v", err)
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Errorf(codes.InvalidArgument, "grpc.Login %v", err)
		}
		return nil, status.Errorf(codes.Internal, "grpc.Login %v", err)
	}
	return &api.LoginResponse{AccessToken: token}, nil
}

func (h *AuthHandler) Register(ctx context.Context, request *api.RegisterRequest) (*api.RegisterResponse, error) {
	token, err := h.as.Register(ctx, request.GetEmail(), request.GetLogin(), request.GetName(), request.GetPassword())
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "grpc.Register %v", err)
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Errorf(codes.InvalidArgument, "grpc.Register %v", err)
		}
		return nil, status.Errorf(codes.Internal, "grpc.Register %v", err)
	}
	return &api.RegisterResponse{AccessToken: token}, nil
}

func (h *AuthHandler) GetUsersInfo(ctx context.Context, request *api.GetUsersInfoRequest) (*api.GetUsersInfoResponse, error) {
	users, err := h.as.GetUsersInfo(ctx, request.GetUserIds())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "grpc.GetUsersInfo %v", err)
		}
		return nil, status.Errorf(codes.Internal, "grpc.GetUsersInfo %v", err)
	}
	usersMap := make(map[string]*api.UserInfo)
	for _, user := range users {
		userInfo := &api.UserInfo{
			Id:    user.ID.String(),
			Name:  user.Name,
			Login: user.Login,
		}
		usersMap[user.ID.String()] = userInfo
	}
	return &api.GetUsersInfoResponse{Users: usersMap}, nil
}

func (h *AuthHandler) GetUserInfoByLogin(ctx context.Context, request *api.GetUserInfoByLoginRequest) (*api.UserInfo, error) {
	user, err := h.as.GetUserInfoByLogin(ctx, request.GetLogin())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "grpc.GetUserInfoByLogin %v", err)
		}
		return nil, status.Errorf(codes.Internal, "grpc.GetUserInfoByLogin %v", err)
	}
	return &api.UserInfo{
		Id:    user.ID.String(),
		Name:  user.Name,
		Login: user.Login,
	}, nil
}

func (h *AuthHandler) GenerateTgLink(ctx context.Context, request *api.GenerateTgLinkRequest) (*api.GenerateTgLinkResponse, error) {
	token, err := h.as.SaveTgToken(ctx, request.GetUserId())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "grpc.GenerateTgLink %v", err)
		}
		return nil, status.Errorf(codes.Internal, "grpc.GenerateTgLink %v", err)
	}
	return &api.GenerateTgLinkResponse{Token: token}, nil
}

func (h *AuthHandler) BindTelegram(ctx context.Context, request *api.BindTelegramRequest) (*api.BindTelegramResponse, error) {
	err := h.as.BindTgUser(ctx, request.GetToken(), request.GetChatId())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "grpc.BindTelegram %v", err)
		}
		return nil, status.Errorf(codes.Internal, "grpc.BindTelegram %v", err)
	}
	return &api.BindTelegramResponse{
		Success: true,
	}, nil
}
