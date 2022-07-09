package gcs

import (
	"context"
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/isd-sgcu/rnkm65-file/src/app/model/file"
	"github.com/isd-sgcu/rnkm65-file/src/config"
	fMock "github.com/isd-sgcu/rnkm65-file/src/mocks/file"
	mock "github.com/isd-sgcu/rnkm65-file/src/mocks/gcs"
	"github.com/isd-sgcu/rnkm65-file/src/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
)

type GCSServiceTest struct {
	suite.Suite
	conf       config.GCS
	filename   string
	filenameIn string
	file       []byte
	url        string
	err        error
	f          *file.File
}

func TestGCSService(t *testing.T) {
	suite.Run(t, new(GCSServiceTest))
}

func (t *GCSServiceTest) SetupTest() {
	t.filenameIn = faker.Word()
	t.filename = fmt.Sprintf("file-%s", t.filenameIn)

	t.url = faker.URL()

	t.conf = config.GCS{
		ProjectId:           faker.Word(),
		BucketName:          faker.Word(),
		Secret:              faker.Word(),
		ServiceAccountKey:   []byte(faker.Word()),
		ServiceAccountEmail: faker.Word(),
	}

	t.f = &file.File{
		Filename: t.filename,
		OwnerID:  faker.UUIDDigit(),
		Tag:      1,
	}

	t.file = []byte("Hello")

	t.err = errors.New("Something wrong :(")
}

func (t *GCSServiceTest) TestUploadImageSuccess() {
	want := &proto.UploadImageResponse{Url: t.url}

	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(nil)
	c.On("GetSignedUrl", t.filename).Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("CreateOrUpdate", t.f).Return(t.f, nil)

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.UploadImage(context.Background(), &proto.UploadImageRequest{
		Filename: t.filename,
		Data:     t.file,
		UserId:   t.f.OwnerID,
		Tag:      1,
	})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *GCSServiceTest) TestUploadImageFailed() {
	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(errors.New("Cannot upload file"))
	c.On("GetSignedUrl", t.filename).Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("CreateOrUpdate", t.f).Return(t.f, nil)

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.UploadImage(context.Background(), &proto.UploadImageRequest{
		Filename: t.filename,
		Data:     t.file,
		Tag:      1,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestUploadImageSaveFileError() {
	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(errors.New("Cannot upload file"))
	c.On("GetSignedUrl", t.filename).Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("CreateOrUpdate", t.f).Return(nil, errors.New("Error while saving"))

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.UploadImage(context.Background(), &proto.UploadImageRequest{
		Filename: t.filename,
		Data:     t.file,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestUploadFileSuccess() {
	want := &proto.UploadFileResponse{Url: t.url}

	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(nil)
	c.On("GetSignedUrl", t.filename).Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("CreateOrUpdate", t.f).Return(t.f, nil)

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.UploadFile(context.Background(), &proto.UploadFileRequest{
		Filename: t.filename,
		Data:     t.file,
		Tag:      1,
	})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *GCSServiceTest) TestUploadFileFailed() {
	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(errors.New("Cannot upload file"))
	c.On("GetSignedUrl", t.filenameIn).Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("CreateOrUpdate", t.f).Return(t.f, nil)

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.UploadFile(context.Background(), &proto.UploadFileRequest{
		Filename: t.filename,
		Data:     t.file,
		Tag:      1,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestUploadFileSaveFileError() {
	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(errors.New("Cannot upload file"))
	c.On("GetSignedUrl", t.filenameIn).Return(t.url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("CreateOrUpdate", t.f).Return(nil, errors.New("Error while saving"))

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.UploadImage(context.Background(), &proto.UploadImageRequest{
		Filename: t.filename,
		Data:     t.file,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestGetSignedUrlSuccess() {
	t.f.Filename = fmt.Sprintf("%s-%s", t.filename, faker.Word())
	str := strings.Split(t.f.Filename, "file-")
	t.filename = str[1]

	url := faker.URL()
	want := &proto.GetSignedUrlResponse{Url: url}

	c := mock.ClientMock{}
	c.On("GetSignedUrl", t.filename).Return(url, nil)

	repo := fMock.RepositoryMock{}
	repo.On("FindByOwnerID", t.f.OwnerID, &file.File{}).Return(t.f, nil)

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *GCSServiceTest) TestGetSignedUrlFailed() {
	t.f.Filename = fmt.Sprintf("%s-%s", t.filename, faker.Word())
	str := strings.Split(t.f.Filename, "file-")
	t.filename = str[1]

	c := mock.ClientMock{}
	c.On("GetSignedUrl", t.filename).Return("", t.err)

	repo := fMock.RepositoryMock{}
	repo.On("FindByOwnerID", t.f.OwnerID, &file.File{}).Return(t.f, nil)

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestGetSignedUrlNotFound() {
	c := mock.ClientMock{}
	c.On("GetSignedUrl", t.filename).Return("", t.err)

	repo := fMock.RepositoryMock{}
	repo.On("FindByOwnerID", t.f.OwnerID, &file.File{}).Return(nil, errors.New("Not found file"))

	srv := NewService(t.conf, &c, &repo)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		UserId: t.f.OwnerID,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}
