#!/bin/sh
# Запускается образом postgres один раз, при создании пустого тома.
set -eu

create_database() {
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
		-v name="$1" -v password="$2" <<-'SQL'
		CREATE ROLE :"name" LOGIN PASSWORD :'password';
		CREATE DATABASE :"name" OWNER :"name";
		REVOKE CONNECT ON DATABASE :"name" FROM PUBLIC;
	SQL
}

create_database graph_api "$GRAPH_API_DB_PASSWORD"
create_database graph_admin_api "$GRAPH_ADMIN_API_DB_PASSWORD"
