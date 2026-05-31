package ygoapi

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

func (c *Client) worker(ctx context.Context, jobs <-chan string, dest string, bar *progressbar.ProgressBar, errorLogger *slog.Logger) error {

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
					if errorLogger != nil {
						errorLogger.Error("failed", "url", job, "error", err)
					}
				}
			}
		}
	}
}
