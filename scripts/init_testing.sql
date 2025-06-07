-- init_db.sql

DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'goauth_test') THEN
      CREATE USER goauth_test WITH PASSWORD 'goauth_test';
   END IF;
END
$do$;

CREATE DATABASE goauth_test WITH OWNER goauth_test ENCODING 'UTF8';

\connect goauth_test

ALTER DATABASE goauth_test SET timezone TO 'UTC';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

GRANT ALL PRIVILEGES ON DATABASE goauth_test TO goauth_test;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO goauth_test;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO goauth_test;
GRANT ALL PRIVILEGES ON SCHEMA public TO goauth_test;
