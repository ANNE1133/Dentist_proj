package service

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOService struct {
	client *minio.Client
	bucket string
}

func NewMinIOService(
	endpoint string,
	accessKey string,
	secretKey string,
	bucket string,
	useSSL bool,
) (*MinIOService, error) {

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOService{client: client, bucket: bucket}, nil
}

func (m *MinIOService) UploadTemporaryImage(
	ctx context.Context,
	reader io.Reader,
	objectName string,
	size int64,
	contentType string,
) error {
	_, err := m.client.PutObject(
		ctx,
		m.bucket,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}

func (m *MinIOService) GeneratePresignedURL(
	ctx context.Context,
	objectName string,
	expiry time.Duration,
) (string, error) {
	presigned, err := m.client.PresignedGetObject(ctx, m.bucket, objectName, expiry, url.Values{})
	if err != nil {
		return "", err
	}
	return presigned.String(), nil
}

// DeleteObject removes an object, used to clean up an image whose DB
// insert (dentist row or embedding row) failed afterward.
func (m *MinIOService) DeleteObject(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
}