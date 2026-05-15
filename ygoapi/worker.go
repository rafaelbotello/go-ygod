package ygoapi

import (
	"context"
	"errors"
	"log"
	"path/filepath"
)

func (c *Client) worker(ctx context.Context, jobs <-chan string, dest string) error {

	for {
		select {
		case <-ctx.Done():
			return nil
		case job, ok := <-jobs:
			if !ok {
				return nil
			}

			fileName := filepath.Base(job)
			destPath := filepath.Join(dest, fileName)

			err := c.DownloadImage(ctx, job, destPath)
			if err != nil {
				if errors.Is(err, ErrRateLimitExceeded) {
					return err
				} else {
					log.Printf("error downloading image %s: %v", fileName, err)
				}
			}
		}
	}
}
