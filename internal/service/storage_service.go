package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gracchi-stdio/castogo/internal/config"
)

type StorageService interface {
	UploadFile(ctx context.Context, file io.Reader, filename string) (string, error)
}

type BunnyStorageService struct {
	Endpoint string
	Password string
}

func NewBunnyStorageService(endpoint, password string) *BunnyStorageService {
	return &BunnyStorageService{
		Endpoint: endpoint,
		Password: password,
	}
}

func (s *BunnyStorageService) UploadFile(ctx context.Context, file io.Reader, filename string) (string, error) {
	endpoint, err := url.Parse(config.Cfg.BunnyStorageEndpoint)
	if err != nil {
		return "", err
	}

	uploadURL := endpoint.JoinPath(filename)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL.String(), file)
	if err != nil {
		return "", fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("AccessKey", config.Cfg.BunnyStoragePassword)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload to bunny: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("upload failed: status %d", resp.StatusCode)
	}

	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	return config.Cfg.StorageCDN + "/" + filename, nil
}
