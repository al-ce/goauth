-- init_db.sql

DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'godiscauth') THEN
      CREATE USER godiscauth WITH PASSWORD 'godiscauth';
   END IF;
END
$do$;

CREATE DATABASE godiscauth WITH OWNER godiscauth ENCODING 'UTF8';

\connect godiscauth

ALTER DATABASE godiscauth SET timezone TO 'UTC';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

GRANT ALL PRIVILEGES ON DATABASE godiscauth TO godiscauth;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO godiscauth;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO godiscauth;
GRANT ALL PRIVILEGES ON SCHEMA public TO godiscauth;
