default:
    @just --list

# watch for changes and run the server
watch:
    CompileDaemon \
    --build="go build -o gofit ./cmd/server/main.go" \
    --command="./gofit"

# just test [path]
test path="":
    #!/usr/bin/env sh
    if [ -z "{{path}}" ]; then
        go test -v -json ./... | gotestfmt
    else
        go test -v -json ./{{path}} | gotestfmt
    fi

