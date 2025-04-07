-- init_db.sql

DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'gofit_test') THEN
      CREATE USER gofit_test WITH PASSWORD 'gofit_test';
   END IF;
END
$do$;

CREATE DATABASE gofit_test WITH OWNER gofit_test ENCODING 'UTF8';

\connect gofit_test

ALTER DATABASE gofit_test SET timezone TO 'UTC';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

GRANT ALL PRIVILEGES ON DATABASE gofit_test TO gofit_test;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gofit_test;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO gofit_test;
GRANT ALL PRIVILEGES ON SCHEMA public TO gofit_test;
