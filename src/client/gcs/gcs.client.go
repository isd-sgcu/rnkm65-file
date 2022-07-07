package gcs

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"github.com/isd-sgcu/rnkm65-file/src/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"time"
)

type Client struct {
	client *storage.Client
	ctx    context.Context
	conf   config.GCS
}

func NewClient() *Client {
	ctx := context.Background()
	client, _ := storage.NewClient(ctx)

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

func (c *Client) Upload(files []byte, filename string) error {
	ctx, cancel := context.WithTimeout(c.ctx, 50*time.Second)
	defer cancel()
	defer c.client.Close()

	buf := bytes.NewBuffer(files)

	wc := c.client.Bucket(c.conf.BucketName).Object(filename).NewWriter(ctx)
	wc.ChunkSize = 0

	if _, err := io.Copy(wc, buf); err != nil {
		return errors.New("Error while uploading the object")
	}

	if err := wc.Close(); err != nil {
		return errors.New("Error while closing the connection")
	}
	log.Info().
		Str("bucket", c.conf.BucketName).
		Msgf("Successfully upload image %v", filename)

	return nil
}

func (c *Client) GetSignedUrl(filename string) (string, error) {
	defer c.client.Close()

	url, err := storage.SignedURL(c.conf.BucketName, filename, &storage.SignedURLOptions{
		GoogleAccessID: c.conf.ServiceAccountEmail,
		PrivateKey:     []byte(c.conf.ServiceAccountKey),
		Method:         "GET",
		Expires:        time.Now().Add(48 * time.Hour),
	})
	if err != nil {
		return "", err
	}

	return url, nil
}
