package service

import (
	"context"
	"io"
	"net/url"

	"github.com/gracchi-stdio/castogo/internal/config"
	bunnystorage "github.com/l0wl3vel/bunny-storage-go-sdk"
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

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	bunnyclient := bunnystorage.NewClient(*endpoint, config.Cfg.BunnyStoragePassword)

	if err := bunnyclient.Upload(filename, content, true); err != nil {
		return "", err
	}

	return config.Cfg.StorageCDN + "/" + filename, nil
}
