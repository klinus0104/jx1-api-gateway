#!/usr/bin/env sh
set -eu
: "${MSSQL_URL:?MSSQL_URL is required}"
echo "Migrations are stored in 001_gm_schema.sql. Apply this file with the deployment's SQL Server migration runner."
echo "This wrapper intentionally does not execute arbitrary SQL from HTTP or application startup."
