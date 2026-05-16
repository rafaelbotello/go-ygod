# go-ygod

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, concurrent, and fault-tolerant Go CLI tool for syncing the entire [YGOPRODeck API](https://ygoprodeck.com/api-guide/) Yu-Gi-Oh! card image database to your local machine.

## Architecture Highlights

This is not a standard synchronous web scraper. `go-ygod` is built using production-grade distributed system patterns to ensure maximum throughput while perfectly respecting third-party API limits.

- **Concurrent Worker Pool:** Utilizes `golang.org/x/sync/errgroup` to manage 20 simultaneous download goroutines with shared context propagation.
- **Strict Rate Limiting:** Implements a Token Bucket algorithm (`golang.org/x/time/rate`) to strictly enforce a 15 request/second limit, preventing IP bans and server strain.
- **Exponential Backoff & Jitter:** Automatically handles network timeouts and 500-level errors by backing off exponentially and adding mathematical jitter to prevent "Thundering Herd" traffic spikes.
- **Idempotent Execution:** Safely resumes interrupted syncs. Uses `os.Stat` to check for existing files, skipping them in O(1) time without wasting API tokens or bandwidth.
- **Graceful Shutdown:** Intercepts `SIGINT` (Ctrl+C) via OS signals and context cancellation. This allows active workers to safely flush file buffers to disk before exiting, preventing data corruption.
- **Thread-Safe Logging:** Captures permanent failures and 404 dead links into `failed_images.txt` using Go's `log` package coupled with `O_APPEND` file modes for lock-free concurrency.
- **Rich Terminal UI:** Features a real-time, ANSI-colored progress bar calculating ETA, completion percentage, and live network requests-per-second.

## Repository Structure

The codebase adheres to idiomatic Go layout standards, heavily decoupling the UI/CLI layer from the core domain logic.

```text
go-ygod/
├── cmd/
│   └── ygod/
│       └── main.go           # Thin Main: Initializes OS context, UI, and Orchestrator
├── ygoapi/
│   ├── client.go         # HTTP Client, Token Limiter, and Core API logic
│   ├── orchestrator.go   # Errgroup Channel management
│   ├── worker.go         # Individual goroutine execution logic
│   └── client_test.go    # Exhaustive unit tests with httptest mock servers
├── images/               # Local cache for downloaded card art (git-ignored)
├── .gitignore
├── go.mod
└── README.md
```
## Installation & Usage

Ensure you have Go 1.22+ installed on your machine.

Clone the repository:

Bash
git clone [https://github.com/YOUR_GITHUB_USERNAME/go-ygod.git](https://github.com/YOUR_GITHUB_USERNAME/go-ygod.git)
cd go-ygod
Download dependencies:

Bash
go mod tidy
Run the data pipeline:

Bash
go run cmd/ygod/main.go
Note: To safely pause the download at any time, press Ctrl+C. The application will finish writing active files and exit gracefully.

## Testing

The ygoapi package includes a robust suite of unit tests. It utilizes net/http/httptest to mock server responses, thoroughly testing rate limiter pacing, context cancellation, and error handling without hitting the live API.

Bash
go test ./ygoapi/ -v

## Disclaimer & API Usage
This tool is designed to strictly adhere to the YGOPRODeck API guidelines (maximum 15 requests per second). Please be a good internet citizen and do not modify the token bucket rate limiter to aggressively spam the API.