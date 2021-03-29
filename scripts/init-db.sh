#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER desafio_go WITH PASSWORD 'desafio_go';
    CREATE DATABASE desafio_go;
    GRANT ALL PRIVILEGES ON DATABASE desafio_go TO desafio_go;
EOSQL
