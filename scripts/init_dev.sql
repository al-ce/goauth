-- init_db.sql

DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'gofit') THEN
      CREATE USER gofit WITH PASSWORD 'gofit';
   END IF;
END
$do$;

CREATE DATABASE gofit WITH OWNER gofit ENCODING 'UTF8';

\connect gofit

ALTER DATABASE gofit SET timezone TO 'UTC';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

GRANT ALL PRIVILEGES ON DATABASE gofit TO gofit;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gofit;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO gofit;
GRANT ALL PRIVILEGES ON SCHEMA public TO gofit;
