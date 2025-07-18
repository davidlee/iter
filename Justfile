
build:
  go build -o vice .

test:
  go test ./...

format:
  gofumpt -l -w .

lint:
  golangci-lint run

lint-single file:
  golangci-lint run {{file}}

run:
  go run . # [subcommand]

wip:
  glow kanban/in-progress

plan:
  glow kanban/backlog

logs:
  uvx claude-code-log --open-browser

clean-test-cache:
  go clean -testcache

devinstall:
  # Error: Wails applications will not build without the correct build tags.
  # go get github.com/go-architect/go-architect@latest
  go get github.com/fdaines/spm-go@latest
  go get github.com/Skarlso/effrit@latest

