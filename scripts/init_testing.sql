-- init_db.sql

DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'godiscauth_test') THEN
      CREATE USER godiscauth_test WITH PASSWORD 'godiscauth_test';
   END IF;
END
$do$;

CREATE DATABASE godiscauth_test WITH OWNER godiscauth_test ENCODING 'UTF8';

\connect godiscauth_test

ALTER DATABASE godiscauth_test SET timezone TO 'UTC';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

GRANT ALL PRIVILEGES ON DATABASE godiscauth_test TO godiscauth_test;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO godiscauth_test;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO godiscauth_test;
GRANT ALL PRIVILEGES ON SCHEMA public TO godiscauth_test;
