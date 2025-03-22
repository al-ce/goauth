default:
    @just --list

# just test [path]
test path="":
    #!/usr/bin/env sh
    if [ -z "{{path}}" ]; then
        go test -v -json ./... | gotestfmt
    else
        go test -v -json ./{{path}} | gotestfmt
    fi
