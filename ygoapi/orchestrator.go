package ygoapi

import (
	"context"
	"sync"
)

func (c *Client) DownloadAllImages(ctx context.Context, urls []string, destDir string, workerCount int) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan string, workerCount)

	var wg sync.WaitGroup

	for range workerCount {
		wg.Go(func() {
			c.worker(ctx, jobs, destDir, cancel)
		})
	}

	for _, url := range urls {
		if ctx.Err() != nil {
			break
		}
		jobs <- url
	}

	close(jobs)

	wg.Wait()
}
