#!/bin/sh

set -eu

create_database() {
    DB_NAME="$1"
    DB_USER="$2"
    DB_PASSWORD="$3"

    role_exists="$(psql \
        -v ON_ERROR_STOP=1 \
        -U "$POSTGRES_USER" \
        -d "$POSTGRES_DB" \
        -tAc "SELECT 1 FROM pg_roles WHERE rolname = '$DB_USER'")"

    if [ "$role_exists" != "1" ]; then
        psql \
            -v ON_ERROR_STOP=1 \
            -U "$POSTGRES_USER" \
            -d "$POSTGRES_DB" \
            -v app_user="$DB_USER" \
            -v app_password="$DB_PASSWORD" <<'SQL'
CREATE ROLE :"app_user" WITH LOGIN PASSWORD :'app_password';
SQL
    fi

    db_exists="$(psql \
        -v ON_ERROR_STOP=1 \
        -U "$POSTGRES_USER" \
        -d "$POSTGRES_DB" \
        -tAc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'")"

    if [ "$db_exists" != "1" ]; then
        psql \
            -v ON_ERROR_STOP=1 \
            -U "$POSTGRES_USER" \
            -d "$POSTGRES_DB" \
            -v app_db="$DB_NAME" \
            -v app_user="$DB_USER" <<'SQL'
CREATE DATABASE :"app_db" OWNER :"app_user";
SQL
    fi
}

create_database \
    "$AUTH_DB_NAME" \
    "$AUTH_DB_USER" \
    "$AUTH_DB_PASSWORD"
