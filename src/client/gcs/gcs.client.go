package gcs

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"github.com/isd-sgcu/rnkm65-file/src/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"io"
	"time"
)

type Client struct {
	conf config.GCS
}

const SignUrlExpiresIn = 15

func NewClient(conf config.GCS) *Client {
	return &Client{
		conf: conf,
	}
}

func (c *Client) Upload(files []byte, filename string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 50*time.Second)
	defer cancel()

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(c.conf.ServiceAccountJSON))
	if err != nil {
		log.Error().
			Err(err).
			Str("service", "file").
			Str("module", "gcs client").
			Msg("Cannot create google cloud storage client")
	}
	defer client.Close()

	buf := bytes.NewBuffer(files)

	wc := client.Bucket(c.conf.BucketName).Object(filename).NewWriter(ctx)
	wc.ChunkSize = 0

	if _, err := io.Copy(wc, buf); err != nil {
		return errors.Wrap(err, "Error while uploading the object")
	}

	if err := wc.Close(); err != nil {
		return errors.Wrap(err, "Error while closing the connection")
	}
	log.Info().
		Str("bucket", c.conf.BucketName).
		Str("service", "file").
		Str("module", "gcs client").
		Msgf("Successfully upload image %v", filename)

	return nil
}

func (c *Client) GetSignedUrl(filename string) (string, error) {
	ops := storage.SignedURLOptions{
		GoogleAccessID: c.conf.ServiceAccountEmail,
		PrivateKey:     c.conf.ServiceAccountKey,
		Method:         "GET",
		Expires:        time.Now().Add(SignUrlExpiresIn * time.Minute),
		Scheme:         storage.SigningSchemeV4,
	}

	url, err := storage.SignedURL(c.conf.BucketName, filename, &ops)
	if err != nil {
		return "", err
	}

	return url, nil
}
