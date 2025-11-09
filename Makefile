# Makefile
.PHONY: start-server start-fetcher

start-server:
	go run ./server

start-fetcher:
	go run . urls.txt 3 5
