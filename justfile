PROJECT := "goauth"
DEV_DB := "goauth"
DEV_USER := "goauth"
DEV_PASS := "goauth"
TEST_DB := "goauth_test"
TEST_USER := "goauth_test"
TEST_PASS := "goauth_test"
HOST := "localhost"
PORT := "5432"
DRIVER := "postgres"
INITDEVDB := "scripts/init_dev.sql"
INITTESTDB := "scripts/init_testing.sql"

default:
    @just --list

# watch for changes and run the server
watch:
    CompileDaemon \
    --build="go build -o goauth ./main.go" \
    --command="./goauth"

# go test {{path}} and format the output
test path="":
    #!/usr/bin/env sh
    if [ -z "{{path}}" ]; then
        go test -v -json ./... | gotestfmt
    else
        go test -v -json ./internal/{{path}} | gotestfmt
    fi

# Initialize development database
init env="":
    #!/usr/bin/env sh
    if [ "{{env}}" = "test" ]; then
        sudo -u postgres psql -f {{INITTESTDB}}
    else
        sudo -u postgres psql -f {{INITDEVDB}}
    fi

# Drop database
drop env="":
    #!/usr/bin/env sh
    if [ "{{env}}" = "test" ]; then
        sudo -u postgres psql -c "DROP DATABASE IF EXISTS {{TEST_DB}};"
    else
        sudo -u postgres psql -c "DROP DATABASE IF EXISTS {{DEV_DB}};"
    fi

# Reset database
reset env="":
    just drop {{env}}
    just init {{env}}

# Open database with rainfrog
rain env="":
    #!/usr/bin/env sh
    if [ "{{env}}" = "test" ]; then
        rainfrog \
          --driver="{{DRIVER}}" \
          --username="{{TEST_USER}}" \
          --host="{{HOST}}" \
          --port="{{PORT}}" \
          --database="{{TEST_DB}}" \
          --password="{{TEST_PASS}}"
    else
        rainfrog \
          --driver="{{DRIVER}}" \
          --username="{{DEV_USER}}" \
          --host="{{HOST}}" \
          --port="{{PORT}}" \
          --database="{{DEV_DB}}" \
          --password="{{DEV_PASS}}"
    fi

# send GET request to ping endpoint
ping:
    curl -X 'GET' -v -s -A '{{PROJECT}} justfile' 'http://localhost:3000/ping'
