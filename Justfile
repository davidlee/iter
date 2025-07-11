
build:
  go build -o iter .

test:
  go test ./..

format:
  gofumpt ./..

lint:
  golangci-lint run

run:
  go run . # [subcommand]
