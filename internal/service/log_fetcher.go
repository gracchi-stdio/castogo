package service

import (
	"bufio"
	"compress/gzip"
	"net/http"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

type BunnyLogFetcher struct {
	accessKey  string
	pullZoneID string
	httpClient *http.Client
}

func NewBunnyLogFetcher(accessKey, pullZoneID string) (*BunnyLogFetcher, error) {

	client := http.Client{}
	return &BunnyLogFetcher{
		accessKey:  accessKey,
		pullZoneID: pullZoneID,
		httpClient: &client,
	}, nil
}

// FetcherEntries downloads a single gzip log file, parse it, returns entries.
// The caller (analytics service) decides WHICH files to fetch.

func (f *BunnyLogFetcher) FetcherEntries(filePath string) ([]domain.RawLogEntry, error) {
	reader, _ := bunnyclient.Download(filePath)
	gz, _ := gzip.NewReader(reader)

	scanner := bufio.NewScanner(gz)
	for scanner.Scan() {
		entry, err := ParseLogEntry(scanner.Text())
	}

	return nil, nil
}
