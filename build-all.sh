#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o ./bin/lunix_x64 ./cmd/main.go
GOOS=linux GOARCH=386 go build -o ./bin/lunix_x32 ./cmd/main.go