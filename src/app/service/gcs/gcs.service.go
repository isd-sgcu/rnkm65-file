package gcs

import (
	"context"
	"fmt"
	"github.com/isd-sgcu/rnkm65-file/src/app/model/file"
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
	conf       config.GCS
	client     IClient
	repository IRepository
}

type IClient interface {
	Upload([]byte, string) error
	GetSignedUrl(string) (string, error)
}

type IRepository interface {
	FindByOwnerID(string, *file.File) error
	CreateOrUpdate(*file.File) error
	Delete(string) error
}

func NewService(conf config.GCS, client IClient, repository IRepository) *Service {
	return &Service{
		conf:       conf,
		client:     client,
		repository: repository,
	}
}

func (s *Service) UploadImage(_ context.Context, req *proto.UploadImageRequest) (*proto.UploadImageResponse, error) {
	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "File cannot be empty")
	}

	filename, err := s.GetObjectName(req.Filename, constant.IMAGE)
	if err != nil {
		log.Error().Err(err).
			Str("service", "file").
			Str("module", "upload image").
			Str("file_name", filename).
			Msg("Invalid file type")
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

	f := &file.File{
		Filename: filename,
		OwnerID:  req.UserId,
		Tag:      int(req.Tag),
	}

	err = s.repository.CreateOrUpdate(f)

	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", filename).
			Str("user_id", req.UserId).
			Msg("Error while saving file data")
		return nil, status.Error(codes.Unavailable, "Internal service error")
	}

	url, err := s.client.GetSignedUrl(filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", filename).
			Str("user_id", req.UserId).
			Msg("Error while trying to get signed url")
		return nil, status.Error(codes.Unavailable, "Internal service error")
	}

	return &proto.UploadImageResponse{Url: url}, nil
}

func (s *Service) UploadFile(_ context.Context, req *proto.UploadFileRequest) (*proto.UploadFileResponse, error) {
	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "File cannot be empty")
	}

	filename, err := s.GetObjectName(req.Filename, constant.FILE)
	if err != nil {
		log.Error().Err(err).
			Str("service", "file").
			Str("module", "upload file").
			Str("method", "GetObjectName").
			Str("file_name", filename).
			Msg("Invalid file type")
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

	url, err := s.client.GetSignedUrl(filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", filename).
			Str("user_id", req.UserId).
			Msg("Error while trying to get signed url")
		return nil, status.Error(codes.Unavailable, "Internal service error")
	}

	return &proto.UploadFileResponse{Url: url}, nil
}

func (s *Service) GetSignedUrl(_ context.Context, req *proto.GetSignedUrlRequest) (*proto.GetSignedUrlResponse, error) {
	f := file.File{}
	err := s.repository.FindByOwnerID(req.UserId, &f)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "get signed url").
			Str("user_id", req.UserId).
			Msg("Error while trying to query data")
		return nil, status.Error(codes.Unavailable, "Internal service error")
	}

	url, err := s.client.GetSignedUrl(f.Filename)
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
