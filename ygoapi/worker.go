package ygoapi

import (
	"context"
	"errors"
	"log"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

func (c *Client) worker(ctx context.Context, jobs <-chan string, dest string, bar *progressbar.ProgressBar) error {

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

			if bar != nil {
				bar.Add(1)
			}

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
