package ygoapi

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func (c *Client) DownloadAllImages(ctx context.Context, urls []string, destDir string, workerCount int) error {

	g, ctx := errgroup.WithContext(ctx)

	jobs := make(chan string, workerCount)

	g.Go(func() error {
		defer close(jobs)

		for _, url := range urls {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case jobs <- url:
			}
		}

		return nil
	})

	for range workerCount {
		g.Go(func() error {
			return c.worker(ctx, jobs, destDir)
		})
	}

	return g.Wait()
}
