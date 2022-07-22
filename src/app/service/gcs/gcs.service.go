package gcs

import (
	"context"
	"github.com/go-redis/redis/v8"
	dto "github.com/isd-sgcu/rnkm65-file/src/app/dto/file"
	model "github.com/isd-sgcu/rnkm65-file/src/app/model/file"
	"github.com/isd-sgcu/rnkm65-file/src/app/utils"
	"github.com/isd-sgcu/rnkm65-file/src/config"
	"github.com/isd-sgcu/rnkm65-file/src/constant/file"
	"github.com/isd-sgcu/rnkm65-file/src/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	conf       config.GCS
	ttl        int
	client     IClient
	repository IRepository
	cacheRepo  ICacheRepository
}

type IClient interface {
	Upload([]byte, string) error
	GetSignedUrl(string) (string, error)
}

type IRepository interface {
	FindByOwnerID(string, *model.File) error
	CreateOrUpdate(*model.File) error
	Delete(string) error
}

type ICacheRepository interface {
	SaveCache(string, interface{}, int) error
	GetCache(string, interface{}) error
}

func NewService(conf config.GCS, ttl int, client IClient, repository IRepository, cacheRepo ICacheRepository) *Service {
	return &Service{
		conf:       conf,
		ttl:        ttl,
		client:     client,
		repository: repository,
		cacheRepo:  cacheRepo,
	}
}

func (s *Service) Upload(_ context.Context, req *proto.UploadRequest) (*proto.UploadResponse, error) {
	if req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "File cannot be empty")
	}

	filename, err := utils.GetObjectName(req.Filename, s.conf.Secret, file.Type(req.Type))
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

	f := &model.File{
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

	cacheFile := dto.CacheFile{
		Url:      url,
		Filename: filename,
	}

	err = s.cacheRepo.SaveCache(req.UserId, &cacheFile, s.ttl)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", filename).
			Str("user_id", req.UserId).
			Interface("cache", cacheFile).
			Msg("Error while connecting to redis server")
		return nil, status.Error(codes.Unavailable, "Error while connecting to redis server")
	}

	return &proto.UploadResponse{Url: url}, nil
}

func (s *Service) GetSignedUrl(_ context.Context, req *proto.GetSignedUrlRequest) (*proto.GetSignedUrlResponse, error) {
	cachedFile := &dto.CacheFile{}
	err := s.cacheRepo.GetCache(req.UserId, cachedFile)
	if err == nil {
		return &proto.GetSignedUrlResponse{Url: cachedFile.Url}, nil
	}

	if err != redis.Nil {
		log.Error().
			Err(err).
			Str("module", "get signed url").
			Str("user_id", req.UserId).
			Msg("Error while connecting to redis server")
		return nil, status.Error(codes.Unavailable, "Error while connecting to redis server")
	}

	f := model.File{}
	err = s.repository.FindByOwnerID(req.UserId, &f)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "get signed url").
			Str("user_id", req.UserId).
			Msg("Error while trying to query data")
		return nil, status.Error(codes.NotFound, "Not found file")
	}

	url, err := s.client.GetSignedUrl(f.Filename)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload image").
			Str("filename", cachedFile.Filename).
			Str("user_id", req.UserId).
			Msg("Cannot connect to google cloud storage")
		return nil, status.Error(codes.Unavailable, "Cannot connect to google cloud storage")
	}

	cachedFile = &dto.CacheFile{
		Url:      url,
		Filename: f.Filename,
	}

	err = s.cacheRepo.SaveCache(req.UserId, cachedFile, s.ttl)
	if err != nil {
		log.Error().
			Err(err).
			Str("module", "upload file").
			Str("filename", cachedFile.Filename).
			Str("user_id", req.UserId).
			Interface("cache", cachedFile).
			Msg("Error while connecting to redis server")
		return nil, status.Error(codes.Unavailable, "Error while connecting to redis server")
	}

	return &proto.GetSignedUrlResponse{Url: url}, nil
}
