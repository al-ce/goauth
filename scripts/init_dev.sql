-- init_db.sql

DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'goauth') THEN
      CREATE USER goauth WITH PASSWORD 'goauth';
   END IF;
END
$do$;

CREATE DATABASE goauth WITH OWNER goauth ENCODING 'UTF8';

\connect goauth

ALTER DATABASE goauth SET timezone TO 'UTC';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

GRANT ALL PRIVILEGES ON DATABASE goauth TO goauth;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO goauth;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO goauth;
GRANT ALL PRIVILEGES ON SCHEMA public TO goauth;
