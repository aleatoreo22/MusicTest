# Variáveis
BINARY_NAME=music-api
MAIN_PATH=cmd/api/main.go

.PHONY: all build run test test-v cover clean tidy help client client-build

all: build

## build: Compila o binário do projeto
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

## run: Inicia o servidor localmente
run:
	go run $(MAIN_PATH)

## test: Executa todos os testes do projeto
test:
	go test ./...

## test-v: Executa testes em modo verboso
test-v:
	go test -v ./...

## cover: Mostra a cobertura de código dos testes
cover:
	go test -cover ./...

## tidy: Limpa e organiza as dependências do go.mod
tidy:
	go mod tidy

## clean: Remove binários e arquivos temporários
clean:
	rm -f $(BINARY_NAME)
	rm -f music.db
	rm -f test.db

## client: Executa o cliente TUI
client:
	go run ./cmd/client

## client-build: Compila o cliente TUI em um binário
client-build:
	go build -o music-client ./cmd/client

## help: Mostra os comandos disponíveis
help:
	@echo "Comandos disponíveis:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' |  sed -e 's/^/ /'
