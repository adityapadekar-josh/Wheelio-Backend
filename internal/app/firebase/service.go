package firebase

import (
	"context"
	"log/slog"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/config"
	"google.golang.org/api/option"
)

type service struct {
	bucket *storage.BucketHandle
}

type Service interface {
	GenerateSignedURL(ctx context.Context, objectPath string, contentType string, expires time.Duration) (string, error)
}

func NewService(bucket *storage.BucketHandle) Service {
	return &service{
		bucket: bucket,
	}
}

func InitFirebaseStorage(ctx context.Context, cfg config.Config) (*storage.BucketHandle, error) {
	opt := option.WithCredentialsFile(cfg.FirebaseService.CredentialsFile)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		slog.Error("failed to initialize firebase app", "error", err)
		return nil, err
	}

	client, err := app.Storage(ctx)
	if err != nil {
		slog.Error("failed to create firebase client", "error", err)

		return nil, err
	}

	bucket, err := client.Bucket(cfg.FirebaseService.BucketName)
	if err != nil {
		slog.Error("failed to initialize firebase bucket", "error", err)
		return nil, err
	}

	err = ConfigureCORS(ctx, bucket)
	if err != nil {
		slog.Error("failed to configure CORS, uploads may not work from browser", "error", err)
		return nil, err
	}

	return bucket, nil
}

func ConfigureCORS(ctx context.Context, bucket *storage.BucketHandle) error {
	cors := []storage.CORS{
		{
			MaxAge:          3600,
			Methods:         []string{"PUT", "POST", "GET", "HEAD", "DELETE", "OPTIONS"},
			Origins:         []string{"*"},
			ResponseHeaders: []string{"Content-Type", "x-goog-meta-*"},
		},
	}

	if _, err := bucket.Update(ctx, storage.BucketAttrsToUpdate{
		CORS: cors,
	}); err != nil {
		slog.Error("failed to set CORS configuration", "error", err)
		return err
	}

	return nil
}

func (s *service) GenerateSignedURL(ctx context.Context, objectPath string, contentType string, expires time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:      "PUT",
		ContentType: contentType,
		Expires:     time.Now().Add(expires),
		Headers: []string{
			"Access-Control-Allow-Origin: *",
			"Access-Control-Allow-Methods: PUT, POST, GET, HEAD, DELETE, OPTIONS",
		},
	}

	url, err := s.bucket.SignedURL(objectPath, opts)
	if err != nil {
		slog.Error("failed to generate signed url for upload", "error", err)
		return "", err
	}

	return url, nil
}
