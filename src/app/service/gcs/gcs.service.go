package gcs

import (
	"context"
	"fmt"
	"github.com/isd-sgcu/rnkm65-file/src/app/utils"
	"github.com/isd-sgcu/rnkm65-file/src/config"
	"github.com/isd-sgcu/rnkm65-file/src/constant"
	"github.com/isd-sgcu/rnkm65-file/src/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type Service struct {
	conf   config.GCS
	client IClient
}

type IClient interface {
	Upload([]byte, string) error
	GetSignedUrl(string) (string, error)
}

func NewService(conf config.GCS, client IClient) *Service {
	return &Service{
		conf:   conf,
		client: client,
	}
}

func (s *Service) UploadImage(_ context.Context, req *proto.UploadImageRequest) (*proto.UploadImageResponse, error) {
	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "File cannot be empty")
	}

	filename, err := s.GetObjectName(req.Filename, constant.IMAGE)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid file type")
	}

	err = s.client.Upload(req.Data, filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Msg("Cannot connect to google cloud storage")
		return nil, status.Error(codes.Unavailable, "Cannot connect to google cloud storage")
	}

	return &proto.UploadImageResponse{Filename: filename}, nil
}

func (s *Service) UploadFile(_ context.Context, req *proto.UploadFileRequest) (*proto.UploadFileResponse, error) {
	filename, err := s.GetObjectName(req.Filename, constant.FILE)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid file type")
	}

	err = s.client.Upload(req.Data, filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Msg("Cannot connect to google cloud storage")
		return nil, status.Error(codes.Unavailable, "Cannot connect to google cloud storage")
	}

	return &proto.UploadFileResponse{Filename: filename}, nil
}

func (s *Service) GetSignedUrl(_ context.Context, req *proto.GetSignedUrlRequest) (*proto.GetSignedUrlResponse, error) {
	url, err := s.client.GetSignedUrl(req.Filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Msg("Cannot connect to google cloud storage")
		return nil, status.Error(codes.Unavailable, "Cannot connect to google cloud storage")
	}

	return &proto.GetSignedUrlResponse{Url: url}, nil
}

func (s *Service) GetObjectName(filename string, fileType constant.FileType) (string, error) {
	text := fmt.Sprintf("%s%s%v", filename, s.conf.Secret, time.Now().Unix())
	hashed := utils.Hash([]byte(text))

	hashed = strings.ReplaceAll(hashed, "/", "")

	switch fileType {
	case constant.FILE:
		return fmt.Sprintf("file-%s-%d-%s", filename, time.Now().Unix(), hashed), nil
	case constant.IMAGE:
		return fmt.Sprintf("image-%s-%d-%s", filename, time.Now().Unix(), hashed), nil
	default:
		return "", nil
	}
}
