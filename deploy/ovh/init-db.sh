#!/bin/sh
set -eu
psql --username postgres --dbname postgres --set ON_ERROR_STOP=1 \
  --set app_password="$TINYCHESS_DB_PASSWORD" <<'SQL'
CREATE ROLE yourmove LOGIN PASSWORD :'app_password';
CREATE DATABASE yourmove OWNER yourmove;
SQL
