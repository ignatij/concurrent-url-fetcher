package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var httpClient = http.Client{}

type Result string

const (
	SUCCESS_RESULT   Result = "success"
	ERROR_RESULT     Result = "error"
	CONTEXT_DEADLINE Result = "context_deadline"
)

type UrlResult struct {
	result          string
	workerThreadIdx int
	resultType      Result
	url             string
}

type WorkerInput struct {
	url     string
	timeout int
}

func workerThread(workerChan chan WorkerInput, resultChan chan UrlResult, idx int) {
	log.Println("started worker with number: ", idx)
	for workerInput := range workerChan {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(workerInput.timeout)*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, workerInput.url, nil)
			if err != nil {
				log.Printf("worker %d: build request: %v", idx, err)
				return
			}
			response, err := httpClient.Do(req)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded {
					resultChan <- UrlResult{
						workerThreadIdx: idx,
						resultType:      CONTEXT_DEADLINE,
						url:             workerInput.url,
					}
					return
				}
				log.Printf("worker %d: request failed: %v", idx, err)
				return
			}
			defer response.Body.Close()
			content, err := io.ReadAll(response.Body)
			if err != nil {
				log.Printf("worker %d: request parsing response failed: %v", idx, err)
				return
			}
			resultType := SUCCESS_RESULT
			if response.StatusCode >= 400 {
				resultType = ERROR_RESULT
			}
			resultChan <- UrlResult{
				result:          string(content),
				workerThreadIdx: idx,
				resultType:      resultType,
				url:             workerInput.url,
			}
		}()

	}
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <url_file> <workers> [timeout_seconds]", os.Args[0])
	}
	urlFile := os.Args[1]
	workers, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("error reading/parsing workers argument")
	}
	workerChan := make(chan WorkerInput, workers)
	resultChan := make(chan UrlResult)
	wg := sync.WaitGroup{}
	for idx := range workers {
		go func(idx int) {
			wg.Go(func() {
				workerThread(workerChan, resultChan, idx+1)
			})
		}(idx)
	}

	timeout, err := strconv.Atoi(os.Args[3])

	if err != nil {
		log.Fatal("error reading/parsing timeout argument")
	}

	f, err := os.Open(urlFile)
	if err != nil {
		log.Fatal("error while opening file: ", urlFile)
	}

	sc := bufio.NewScanner(f)
	go func() {
		for sc.Scan() {
			workerChan <- WorkerInput{sc.Text(), timeout}
		}
		close(workerChan)
	}()

	go func() {
		// close the result channel once all worker threads are finished with the work
		wg.Wait()
		close(resultChan)
	}()

	for r := range resultChan {
		log.Printf("URL: %s run on worker %d with result type of %s and result: %s\n", r.url, r.workerThreadIdx, r.resultType, r.result)
	}
}
