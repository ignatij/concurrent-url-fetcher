# Concurrent URL Fetcher

## Prerequisites
- Go 1.25+ available in your PATH

## Run the local test server
Use the provided Makefile target to start the HTTP server that exposes the `/fast`, `/slow`, and `/error` endpoints described in `server/server.go`:

```bash
make start-server
```

Leave this process running in its terminal so the fetcher has something to hit.

## Run the concurrent fetcher
With the server running, launch the fetcher from a new terminal using the second Makefile target:

```bash
make start-fetcher
```

This command executes `go run . urls.txt 3 5`, which means:
- `urls.txt` is the input file containing the list of URLs to fetch.
- `3` is the number of worker goroutines that will process URLs concurrently.
- `5` is the per-request timeout in seconds; any URL taking longer than five seconds is canceled.

Adjust `urls.txt` or the Makefile target if you want to test different inputs, worker counts, or timeouts.
