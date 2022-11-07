# Snort3 Parser

Collect the Snort v3 alert log and send it to Mata Elang Defense Center through MQTT

## Requirements
 - [Golang](https://go.dev/dl)
 - make (linux)

## Build with Makefile
 - Show the make help: `make help`
 - Build the binary: `make build`
 - The binary will be available at `out/bin` directory.

## Run using go command
 - `go mod download`
 - `go run main.go`
