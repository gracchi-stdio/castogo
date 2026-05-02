package service

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gracchi-stdio/castogo/internal/domain"
)

type LogFetcher interface {
	FetchEntries(ctx context.Context, t time.Time) ([]domain.RawLogEntry, error)
}

type BunnyLogFetcher struct {
	baseURL    string // "https://logging.bunnycdn.com" in prod, test server URL in tests
	accessKey  string
	pullZoneID string
	httpClient *http.Client
}

func NewBunnyLogFetcher(baseURL, accessKey, pullZoneID string, httpClient *http.Client) *BunnyLogFetcher {
	return &BunnyLogFetcher{
		baseURL:    baseURL,
		accessKey:  accessKey,
		pullZoneID: pullZoneID,
		httpClient: httpClient,
	}
}

func (f *BunnyLogFetcher) FetchEntries(ctx context.Context, t time.Time) ([]domain.RawLogEntry, error) {
	timeStr := t.Format("01-02-06")
	url := fmt.Sprintf("%s/%s/%s.log", f.baseURL, timeStr, f.pullZoneID)

	fmt.Printf("%s[analytics]%s fetching %s\n", cCyan, cReset, url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return nil, fmt.Errorf("fetch logs %s: %s", url, resp.Status)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	return ParseLogEntries(gzReader)
}
