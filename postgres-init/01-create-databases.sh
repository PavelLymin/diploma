set -e

createdb \
    --username "$POSTGRES_USER" \
    --maintenance-db "$POSTGRES_DB" \
    "$AUTH_DB"

createdb \
    --username "$POSTGRES_USER" \
    --maintenance-db "$POSTGRES_DB" \
    "$USERS_DB"