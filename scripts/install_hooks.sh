#!/usr/bin/env bash
cd "$(git rev-parse --show-toplevel)" || exit 1
cd .git/hooks || exit 1

cat << 'EOF' > pre-commit
#!/bin/sh
cd "$(git rev-parse --show-toplevel)" || exit 1

docker run --rm -d --name goauth-test-db \
    -e POSTGRES_DB=goauth_test \
    -e POSTGRES_USER=goauth_test \
    -e POSTGRES_PASSWORD=goauth_test \
    -p 5433:5432 postgres:15

sleep 3

PGHOST=localhost PGPORT=5433 PGUSER=goauth_test PGPASSWORD=goauth_test \
    psql -f scripts/init_testing.sql || exit 1

go test ./...
TEST_RESULT=$?

docker stop goauth-test-db

[ $TEST_RESULT -ne 0 ] && exit 1
exit 0
EOF
chmod +x pre-commit
