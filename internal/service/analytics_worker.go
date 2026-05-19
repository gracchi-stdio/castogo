package service

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	cReset = "\033[0m"
	cCyan  = "\033[36m"
	cDim   = "\033[2m"
)

type AnalyticsWorker struct {
	service  *AnalyticsService
	interval time.Duration
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewAnalyticsWorker(service *AnalyticsService, interval time.Duration) *AnalyticsWorker {
	return &AnalyticsWorker{
		service:  service,
		interval: interval,
	}
}

func (w *AnalyticsWorker) Start(ctx context.Context) {
	ctx, w.cancel = context.WithCancel(ctx)

	// Tell WaitGroup that we're starting a new goroutine
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		fmt.Printf("%s[analytics]%s worker started — interval: %s\n", cCyan, cReset, w.interval)

		// Process once immediately on start
		if err := w.service.ProcessLogsForDate(ctx, yesterday()); err != nil {
			fmt.Printf("%s[analytics]%s error processing logs: %s%v%s\n", cCyan, cReset, cDim, err, cReset)
		}

		for {
			select {
			case <-ticker.C:
				date := yesterday()
				fmt.Printf("%s[analytics]%s tick — processing %s\n", cCyan, cReset, date.Format("2006-01-02"))
				if err := w.service.ProcessLogsForDate(ctx, date); err != nil {
					fmt.Printf("%s[analytics]%s error processing %s: %s%v%s\n", cCyan, cReset, date.Format("2006-01-02"), cDim, err, cReset)
				}
			case <-ctx.Done():
				fmt.Printf("%s[analytics]%s worker stopping...\n", cCyan, cReset)
				return
			}
		}
	}()
}

func (w *AnalyticsWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func yesterday() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.AddDate(0, 0, -1).Day(), 0, 0, 0, 0, now.Location())
}
