package service

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"time"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

type BunnyLogFetcher struct {
	accessKey  string
	pullZoneID string
	httpClient *http.Client
}

func NewBunnyLogFetcher(accessKey, pullZoneID string, httpClient *http.Client) (*BunnyLogFetcher, error) {

	return &BunnyLogFetcher{
		accessKey:  accessKey,
		pullZoneID: pullZoneID,
		httpClient: &client,
	}, nil
}

// FetcherEntries downloads a single gzip log file, parse it, returns entries.
// The caller (analytics service) decides WHICH files to fetch.
// use gzip

func (f *BunnyLogFetcher) FetcherEntries(filePath string) ([]domain.RawLogEntry, error) {
	url := fmt.Sprintf("https://logging.bunnycdn.com/%s/%s", time.Now().Format("{MM}-{DD}-{YY}"), f.pullZoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("AccessKey", f.accessKey)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch log file: %s", resp.Status)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	return nil, nil
}
