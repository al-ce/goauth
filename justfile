# justfile docs https://just.systems/man/en/
# This file is used for local development convenience
#
# Dependencies:
# - just https://github.com/casey/just?tab=readme-ov-file
# - psql https://archlinux.org/packages/?name=postgresql
# - rainfrog https://github.com/achristmascarl/rainfrog
# - gotestfmt https://github.com/GoTestTools/gotestfmt
# - CompileDaemon https://github.com/githubnemo/CompileDaemon

set quiet := true

PROJECT := "godiscauth"

# credentials

DEV_DB := "godiscauth"
DEV_USER := "godiscauth"
DEV_PASS := "godiscauth"
TEST_DB := "goauth_test"
TEST_USER := "goauth_test"
TEST_PASS := "goauth_test"
def_user := "bob@bob.com"
def_pw := "89532c353a03603b48b2dfe3ea21738b63560be53f63d551b1587cb75f997b36"

# pg url

DB_PORT := "5432"
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
    just request deleteaccount DELETE "{}" {{ email }} && \\
    rm /tmp/goauth_curl_justfile_cookies/{{ email }} # clean up cookie

# #############################################################################
# Development
# #############################################################################

# Run the application and watch for changes, recompile/restart on changes
[group('dev')]
watch:
    #!/usr/bin/env sh
    export SESSION_KEY=$(date | sha256sum | cut -d' ' -f1)
    export AUTH_SERVER_PORT={{ PORT }}
    export DATABASE_URL="{{ DRIVER }}://{{ DEV_USER }}:{{ DEV_PASS }}@{{ HOST }}:{{ DB_PORT }}/{{ DEV_DB }}"
    echo "Using DB: $DB"
    CompileDaemon \
    --build="go build -o {{ PROJECT }} ./main.go" \
    --command="./{{ PROJECT }}"

# go test {{path}} and format the output
[group('dev')]
test path="":
    #!/usr/bin/env sh
    just init test
    if [ -z "{{ path }}" ]; then
        go test -v -json ./... | gotestfmt -hide successful-tests
    else
        go test -v -json ./internal/{{ path }} | gotestfmt -hide successful-tests
    fi

# Initialize auth development database
[group('dev')]
init env="":
    #!/usr/bin/env sh
    if [ "{{ env }}" = "test" ]; then
        sudo -u postgres psql -f {{ INITTESTDB }}
    else
        sudo -u postgres psql -f {{ INITDEVDB }}
    fi

# Drop database
[group('dev')]
drop env="":
    #!/usr/bin/env sh
    if [ "{{ env }}" = "test" ]; then
        sudo -u postgres psql -c "DROP DATABASE IF EXISTS {{ TEST_DB }};"
    else
        sudo -u postgres psql -c "DROP DATABASE IF EXISTS {{ DEV_DB }};"
    fi

# Reset database
[group('dev')]
reset env="":
    just drop {{ env }}
    just init {{ env }}

# Open database with rainfrog
[group('dev')]
rain env="":
    #!/usr/bin/env sh
    if [ "{{ env }}" = "test" ]; then
        rainfrog \
          --driver="{{ DRIVER }}" \
          --username="{{ TEST_USER }}" \
          --host="{{ HOST }}" \
          --port="{{ DB_PORT }}" \
          --database="{{ TEST_DB }}" \
          --password="{{ TEST_PASS }}"
    else
        rainfrog \
          --driver="{{ DRIVER }}" \
          --username="{{ DEV_USER }}" \
          --host="{{ HOST }}" \
          --port="{{ DB_PORT }}" \
          --database="{{ DEV_DB }}" \
          --password="{{ DEV_PASS }}"
    fi

# Connect to database with psql
[group('dev')]
pg env="":
    #!/usr/bin/env sh
    if [ "{{ env }}" = "test" ]; then
        psql -h {{ HOST }} -p {{ DB_PORT }} -U {{ TEST_USER }} {{ TEST_DB }}
    else
        psql -h {{ HOST }} -p {{ DB_PORT }} -U {{ DEV_USER }} {{ DEV_DB }}
    fi
