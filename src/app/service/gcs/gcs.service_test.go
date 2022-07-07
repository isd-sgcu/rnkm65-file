package gcs

import (
	"context"
	"github.com/bxcodec/faker/v3"
	"github.com/isd-sgcu/rnkm65-file/src/config"
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
	conf     config.GCS
	filename string
	file     []byte
	err      error
}

func TestGCSService(t *testing.T) {
	suite.Run(t, new(GCSServiceTest))
}

func (t *GCSServiceTest) SetupTest() {
	t.filename = faker.Word()

	t.conf = config.GCS{
		ProjectId:           faker.Word(),
		BucketName:          faker.Word(),
		Secret:              faker.Word(),
		ServiceAccountKey:   faker.Word(),
		ServiceAccountEmail: faker.Word(),
	}

	t.file = []byte("Hello")

	t.err = errors.New("Something wrong :(")
}

func (t *GCSServiceTest) TestUploadImageSuccess() {
	want := t.filename

	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(nil)

	srv := NewService(t.conf, &c)

	actual, err := srv.UploadImage(context.Background(), &proto.UploadImageRequest{
		Filename: t.filename,
		Data:     t.file,
	})

	names := strings.Split(actual.Filename, "-")

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, names[1])
}

func (t *GCSServiceTest) TestUploadImageFailed() {
	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(errors.New("Cannot upload file"))

	srv := NewService(t.conf, &c)

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
	want := t.filename

	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(nil)

	srv := NewService(t.conf, &c)

	actual, err := srv.UploadFile(context.Background(), &proto.UploadFileRequest{
		Filename: t.filename,
		Data:     t.file,
	})

	names := strings.Split(actual.Filename, "-")

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, names[1])
}

func (t *GCSServiceTest) TestUploadFileFailed() {
	c := mock.ClientMock{}
	c.On("Upload", t.file).Return(errors.New("Cannot upload file"))

	srv := NewService(t.conf, &c)

	actual, err := srv.UploadFile(context.Background(), &proto.UploadFileRequest{
		Filename: t.filename,
		Data:     t.file,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *GCSServiceTest) TestGetSignedUrlSuccess() {
	url := faker.URL()
	want := &proto.GetSignedUrlResponse{Url: url}

	c := mock.ClientMock{}
	c.On("GetSignedUrl", t.filename).Return(url, nil)

	srv := NewService(t.conf, &c)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		Filename: t.filename,
	})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *GCSServiceTest) TestGetSignedUrlFailed() {
	c := mock.ClientMock{}
	c.On("GetSignedUrl", t.filename).Return("", t.err)

	srv := NewService(t.conf, &c)

	actual, err := srv.GetSignedUrl(context.Background(), &proto.GetSignedUrlRequest{
		Filename: t.filename,
	})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}
