#!/usr/bin/env bash
# load_synthea.sh — Loads Synthea sample CSV data into the running MySQL container.
#
# Usage:
#   ./load_synthea.sh [CSV_DIR] [DSN]
#
# Defaults:
#   CSV_DIR = ./csv  (relative to this script's location)
#   DSN     = cosmo:cosmo@tcp(localhost:3306)/cosmo_db
#
# Prerequisites:
#   - MySQL container must be running (docker compose up -d)
#   - mysql CLI must be installed locally
#   - LOAD DATA LOCAL INFILE must be enabled (this script enables it per-session)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CSV_DIR="${1:-${SCRIPT_DIR}/csv}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-cosmo}"
DB_PASS="${DB_PASS:-cosmo}"
DB_NAME="${DB_NAME:-cosmo_db}"

MYSQL_CMD="mysql --local-infile=1 -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASS} ${DB_NAME}"

echo "▶ Loading Synthea CSV data from: ${CSV_DIR}"
echo "▶ Target: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo

load_table() {
    local TABLE="$1"
    local FILE="${CSV_DIR}/${TABLE}.csv"

    if [[ ! -f "${FILE}" ]]; then
        echo "  ⚠ Skipping ${TABLE}: file not found at ${FILE}"
        return
    fi

    echo -n "  → Loading ${TABLE}... "
    ${MYSQL_CMD} --execute="
        LOAD DATA LOCAL INFILE '${FILE}'
        INTO TABLE \`${TABLE}\`
        FIELDS TERMINATED BY ','
        OPTIONALLY ENCLOSED BY '\"'
        LINES TERMINATED BY '\n'
        IGNORE 1 ROWS;
    "
    ROW_COUNT=$(${MYSQL_CMD} --skip-column-names --execute="SELECT COUNT(*) FROM \`${TABLE}\`;")
    echo "done (${ROW_COUNT} rows)"
}

load_table patients
load_table organizations
load_table payers
load_table providers
load_table conditions
load_table medications
load_table encounters
load_table observations
load_table procedures
load_table allergies
load_table careplans
load_table devices
load_table imaging_studies
load_table immunizations
load_table payer_transitions
load_table supplies

echo
echo "✅ All tables loaded. Run the MCP server with:"
echo "   ./schema-mcp --dsn \"${DB_USER}:${DB_PASS}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}\" \\"
echo "                --context $(dirname ${SCRIPT_DIR})/synthea/context.yaml"
