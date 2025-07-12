
build:
  go build -o iter .

test:
  go test ./...

format:
  gofumpt -l -w .

lint:
  golangci-lint run

run:
  go run . # [subcommand]
