default:
    @just --list

# watch for changes and run the server
watch:
    CompileDaemon \
    --build="go build -o gofit ./cmd/server/main.go" \
    --command="./gofit"

# go test {{path}} and format the output
test path="":
    #!/usr/bin/env sh
    if [ -z "{{path}}" ]; then
        go test -v -json ./... | gotestfmt
    else
        go test -v -json ./internal/{{path}} | gotestfmt
    fi

# Initialize development database
init-db:
    sudo -u postgres psql -f scripts/init_prod.sql

# Initialize test database
init-test-db:
    sudo -u postgres psql -f scripts/init_testing.sql

# Drop development database
drop-db:
    sudo -u postgres psql -c "DROP DATABASE IF EXISTS gofit;"

# Drop test database
drop-test-db:
    sudo -u postgres psql -c "DROP DATABASE IF EXISTS gofit_test;"

# Reset development database
reset-db: drop-db init-db

# Reset test database
reset-test-db: drop-test-db init-test-db
