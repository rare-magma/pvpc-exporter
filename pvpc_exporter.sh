#!/usr/bin/env bash

set -Eeo pipefail

dependencies=(curl date gzip jq)
for program in "${dependencies[@]}"; do
    command -v "$program" >/dev/null 2>&1 || {
        echo >&2 "Couldn't find dependency: $program. Aborting."
        exit 1
    }
done

if [[ "${RUNNING_IN_DOCKER}" ]]; then
    source "/app/pvpc_exporter.conf"
else
    # shellcheck source=/dev/null
    source "$CREDENTIALS_DIRECTORY/creds"
fi

[[ -z "${INFLUXDB_HOST}" ]] && echo >&2 "INFLUXDB_HOST is empty. Aborting" && exit 1
[[ -z "${INFLUXDB_API_TOKEN}" ]] && echo >&2 "INFLUXDB_API_TOKEN is empty. Aborting" && exit 1
[[ -z "${ORG}" ]] && echo >&2 "ORG is empty. Aborting" && exit 1
[[ -z "${BUCKET}" ]] && echo >&2 "BUCKET is empty. Aborting" && exit 1

CURL=$(command -v curl)
DATE=$(command -v date)
GZIP=$(command -v gzip)
JQ=$(command -v jq)

NB_DAYS="0"

if [[ $# -eq 1 ]]; then
    # regexp for integer assertion
    if [[ $1 =~ ^[0-9]+$ ]]; then
        NB_DAYS="$1"
    else
        echo >&2 "First argument is not a number. Aborting" && exit 1
    fi
fi

RFC_DATE=$($DATE --rfc-3339=date --date="$NB_DAYS days ago")
INFLUXDB_URL="https://$INFLUXDB_HOST/api/v2/write?precision=s&org=$ORG&bucket=$BUCKET"
PVPC_URL="https://apidatos.ree.es/es/datos/mercados/precios-mercados-tiempo-real?"
PVPC_URL+="start_date=${RFC_DATE}T00:00&end_date=${RFC_DATE}T23:59&time_trunc=hour"

pvpc_json=$($CURL --silent --fail --show-error --compressed "$PVPC_URL")

parsed_prices_json=$(
    echo "$pvpc_json" |
        $JQ '.included[] | select(.id=="1001") | del(.attributes.values[].percentage) | .attributes.values'
)

length=$(echo "$parsed_prices_json" | $JQ 'length - 1')

for i in $(seq 0 "$length"); do

    mapfile -t parsed_priced_stats < <(echo "$parsed_prices_json" | $JQ --raw-output ".[$i] | .value,.datetime")
    pvpc_price_value=${parsed_priced_stats[0]}
    ts=$($DATE "+%s" --date="${parsed_priced_stats[1]}")

    price_stats+=$(printf "\npvpc_price price=%s %s" "$pvpc_price_value" "$ts")
done

echo "$price_stats" | $GZIP |
    $CURL --silent --fail --show-error \
        --request POST "${INFLUXDB_URL}" \
        --header 'Content-Encoding: gzip' \
        --header "Authorization: Token $INFLUXDB_API_TOKEN" \
        --header "Content-Type: text/plain; charset=utf-8" \
        --header "Accept: application/json" \
        --data-binary @-
