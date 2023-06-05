LISTEN_ADDR="$1"
SQL_ADDR="$2"
DATA_SRC=/docker-entrypoint-initdb.d

/cockroach/cockroach.sh init --host "$LISTEN_ADDR" --insecure

for sql in "$DATA_SRC"/*.sql; do
   cat "$sql" | cockroach sql --host "$SQL_ADDR" --insecure
done
