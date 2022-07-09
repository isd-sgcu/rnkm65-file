package gcs

import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"strings"
)

type ClientMock struct {
	mock.Mock
}

func (c *ClientMock) Upload(file []byte, _ string) error {
	args := c.Called(file)

	return args.Error(0)
}

func (c *ClientMock) GetSignedUrl(filename string) (string, error) {
	filenames := strings.Split(filename, "-")
	filename = fmt.Sprintf("%s-%s", filenames[1], filenames[2])

	args := c.Called(filename)

	return args.String(0), args.Error(1)
}
