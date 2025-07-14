
build:
  go build -o vice .

test:
  go test ./...

format:
  gofumpt -l -w .

lint:
  golangci-lint run

run:
  go run . # [subcommand]

wip:
  glow kanban/in-progress

plan:
  glow kanban/backlog

logs:
  uvx claude-code-log --open-browser
