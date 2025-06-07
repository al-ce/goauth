# justfile docs https://just.systems/man/en/
# This file is used for local development convenience
#
# Dependencies:
# - just https://github.com/casey/just?tab=readme-ov-file
# - docker or podman
# - psql https://archlinux.org/packages/?name=postgresql
# - rainfrog https://github.com/achristmascarl/rainfrog
# - gotestfmt https://github.com/GoTestTools/gotestfmt
# - CompileDaemon https://github.com/githubnemo/CompileDaemon

set quiet := true

PROJECT := "goauth"

# credentials

DEV_DB := "goauth"
DEV_USER := "goauth"
DEV_PASS := "goauth"
TEST_DB := "goauth_test"
TEST_USER := "goauth_test"
TEST_PASS := "goauth_test"
def_user := "bob@bob.com"
def_pw := "89532c353a03603b48b2dfe3ea21738b63560be53f63d551b1587cb75f997b36"

# Docker container names

DEV_CONTAINER := "goauth-dev-db"
TEST_CONTAINER := "goauth-test-db"

# pg url

DEV_DB_PORT := "5434"
TEST_DB_PORT := "5433"
DRIVER := "postgres"

# Service url

HOST := "localhost"
PORT := "3001"
SERVICE_URL := "localhost:3001"

# Scripts

INITDEVDB := "scripts/init_dev.sql"
INITTESTDB := "scripts/init_testing.sql"

# Other

cookies_dir := "/tmp/goauth_curl_justfile_cookies/"

default:
    @just --list

# #############################################################################
# Direct requests/queries
# #############################################################################

# Make an http request
[group('exec')]
request endpoint method data="" cookies_id="nil":
    mkdir -p {{ cookies_dir }}
    # Get cookies if they exist, also write cookies
    curl -s --request "$(echo {{ method }} | tr [:lower:] [:upper:])"  \
        --data "{{ data }}" \
        --header 'Content-Type: application/json' {{ SERVICE_URL }}/{{ endpoint }} \
        -b {{ cookies_dir }}{{ cookies_id }}      \
        -c {{ cookies_dir }}{{ cookies_id }} | jq

# #############################################################################
# API endpoints
# #############################################################################

# `goauth/ping`
[group('authapi')]
ping:
    curl -X 'GET' -v -s -A '{{ PROJECT }} justfile' 'http://localhost:{{ PORT }}/ping'

# `goauth/register`
[group('authapi')]
register email=(def_user) pw=(def_pw):
    just request register \
    POST '{\"email\":\"{{ email }}\", \"password\":\"{{ pw }}\"}'

# `goauth/login`
[group('authapi')]
login email=(def_user) pw=(def_pw):
    just request login \
    POST '{\"email\":\"{{ email }}\", \"password\":\"{{ pw }}\"}' \
    {{ email }}

# `goauth/whoami`
[group('authapi')]
whoami email=(def_user):
    just request whoami GET "{}" {{ email }}

# `goauth/logout`
[group('authapi')]
logout email=(def_user):
    just request logout POST "{}" {{ email }}

# `goauth/logouteverywhere`
[group('authapi')]
logouteverywhere email=(def_user):
    just request logouteverywhere POST "{}" {{ email }}

# `goauth/updateuser`
[group('authapi')]
updateuser email=(def_user) new_email=(def_user) new_pw=(def_pw):
    just request updateuser POST \
        '{\"email\":\"{{ new_email }}\", \"password\":\"{{ new_pw }}\"}' \
        {{ email }}

# `goauth/deleteaccount`
[group('authapi')]
deleteaccount email=(def_user):
    just request deleteaccount DELETE "{}" {{ email }} && \
    rm /tmp/goauth_curl_justfile_cookies/{{ email }} # clean up cookie

# #############################################################################
# Development
# #############################################################################

# Run the application and watch for changes, recompile/restart on changes
[group('dev')]
watch:
    #!/usr/bin/env sh
    just init
    export SESSION_KEY=$(date | sha256sum | cut -d' ' -f1)
    export AUTH_SERVER_PORT={{ PORT }}
    export DATABASE_URL="{{ DRIVER }}://{{ DEV_USER }}:{{ DEV_PASS }}@{{ HOST }}:{{ DEV_DB_PORT }}/{{ DEV_DB }}"
    echo "Using DB: $DATABASE_URL"
    CompileDaemon \
    --build="go build -o {{ PROJECT }} ./main.go" \
    --command="./{{ PROJECT }}"

# go test {{path}} and format the output
[group('dev')]
test path="":
    #!/usr/bin/env sh
    just init test
    export DATABASE_URL="{{ DRIVER }}://{{ TEST_USER }}:{{ TEST_PASS }}@{{ HOST }}:{{ TEST_DB_PORT }}/{{ TEST_DB }}"
    go clean -testcache | exit 1
    if [ -z "{{ path }}" ]; then
        go test -p 1 -v -json ./... | gotestfmt
    else
        go test -v -json ./internal/{{ path }} | gotestfmt -hide successful-tests
    fi
    TEST_RESULT=$?
    just stop-test-db
    exit $TEST_RESULT

# Initialize database with schema
[group('dev')]
init env="":
    #!/usr/bin/env sh
    if [ "{{ env }}" = "test" ]; then
        # Ensure the test-db is stopped and removed for a clean start
        just stop-test-db
        just start-test-db
        PGHOST=localhost PGPORT={{ TEST_DB_PORT }} PGUSER={{ TEST_USER }} PGPASSWORD={{ TEST_PASS }} \
        psql -f {{ INITTESTDB }}
    else
        # dev-db should persist until stopped
        just start-dev-db
        PGHOST=localhost PGPORT={{ DEV_DB_PORT }} PGUSER={{ DEV_USER }} PGPASSWORD={{ DEV_PASS }} \
        psql -f {{ INITDEVDB }}
    fi

# Open database with rainfrog
[group('dev')]
rain env="":
    #!/usr/bin/env sh
    if [ "{{ env }}" = "test" ]; then
        just start-test-db
        rainfrog \
          --driver="{{ DRIVER }}" \
          --username="{{ TEST_USER }}" \
          --host="{{ HOST }}" \
          --port="{{ TEST_DB_PORT }}" \
          --database="{{ TEST_DB }}" \
          --password="{{ TEST_PASS }}"
    else
        just start-dev-db
        rainfrog \
          --driver="{{ DRIVER }}" \
          --username="{{ DEV_USER }}" \
          --host="{{ HOST }}" \
          --port="{{ DEV_DB_PORT }}" \
          --database="{{ DEV_DB }}" \
          --password="{{ DEV_PASS }}"
    fi

# Connect to database with psql
[group('dev')]
pg env="":
    #!/usr/bin/env sh
    if [ "{{ env }}" = "test" ]; then
        just start-test-db
        PGHOST={{ HOST }} PGPORT={{ TEST_DB_PORT }} PGUSER={{ TEST_USER }} PGPASSWORD={{ TEST_PASS }} \
        psql {{ TEST_DB }}
    else
        just start-dev-db
        PGHOST={{ HOST }} PGPORT={{ DEV_DB_PORT }} PGUSER={{ DEV_USER }} PGPASSWORD={{ DEV_PASS }} \
        psql {{ DEV_DB }}
    fi

# #############################################################################
# Docker Database Management
# #############################################################################

# Start development database container
[group('db')]
start-dev-db:
    #!/usr/bin/env sh
    if ! docker ps --format json | jq -r .Names | grep -q "^ DEV_CONTAINER $"; then
        echo "Starting development database..."
        docker run --rm -d --name {{ DEV_CONTAINER }} \
            -e POSTGRES_DB={{ DEV_DB }} \
            -e POSTGRES_USER={{ DEV_USER }} \
            -e POSTGRES_PASSWORD={{ DEV_PASS }} \
            -p {{ DEV_DB_PORT }}:5432 \
            postgres:15
        sleep 3
        echo "Development database started on port {{ DEV_DB_PORT }}"
    else
        echo "Development database already running"
    fi

# Start test database container
[group('db')]
start-test-db:
    #!/usr/bin/env sh
    if ! docker ps --format json | jq -r .Names | grep -q "^{{ TEST_CONTAINER }}$"; then
        echo "Starting test database..."
        docker run --rm -d --name {{ TEST_CONTAINER }} \
            -e POSTGRES_DB={{ TEST_DB }} \
            -e POSTGRES_USER={{ TEST_USER }} \
            -e POSTGRES_PASSWORD={{ TEST_PASS }} \
            -p {{ TEST_DB_PORT }}:5432 \
            postgres:15
        sleep 3
        echo "Test database started on port {{ TEST_DB_PORT }}"
    else
        echo "Test database already running"
    fi

# Stop development database container
[group('db')]
stop-dev-db:
    #!/usr/bin/env sh
    if ! docker ps --format json | jq -r .Names | grep -q "^{{ DEV_CONTAINER }}$"; then
        echo "Stopping development database..."
        docker stop {{ DEV_CONTAINER }}
        echo "Development database stopped"
    else
        echo "Development database not running"
    fi

# Stop test database container
[group('db')]
stop-test-db:
    #!/usr/bin/env sh
    if docker ps --format json | jq -r .Names | grep -q "^{{ TEST_CONTAINER }}$"; then
        echo "Stopping test database..."
        docker stop {{ TEST_CONTAINER }}
        echo "Test database stopped"
    else
        echo "Test database not running"
    fi

# Stop all database containers
[group('db')]
stop-all-db:
    just stop-dev-db
    just stop-test-db

# Show database container status
[group('db')]
db-status:
    #!/usr/bin/env sh
    echo "Database container status:"
    echo "========================="
    if docker ps --format json | jq -r .Names | grep -q "^{{ DEV_CONTAINER }}$"; then
        echo "Development DB: ✓ Running (port {{ DEV_DB_PORT }})"
    else
        echo "Development DB: ✗ Stopped"
    fi
    if docker ps --format json | jq -r .Names | grep -q "^{{ TEST_CONTAINER }}$"; then
        echo "Test DB:        ✓ Running (port {{ TEST_DB_PORT }})"
    else
        echo "Test DB:        ✗ Stopped"
    fi
