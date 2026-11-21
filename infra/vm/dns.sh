#!/usr/bin/env bash
set -euo pipefail

IP=${1:?usage: $0 <ip>}
API=https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records

cf() {
  curl -sS --fail-with-body --retry 5 --connect-timeout 5 --max-time 30 \
    -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" "$@"
}

for name in "$DOMAIN" "*.$DOMAIN"; do
  id=$(cf -G "$API" --data-urlencode type=A --data-urlencode "name=$name" | jq -r '.result[0].id // empty')
  record=$(jq -nc --arg name "$name" --arg ip "$IP" '{type: "A", name: $name, content: $ip, ttl: 60, proxied: false}')
  if [[ -n $id ]]; then
    cf -X PUT "$API/$id" --json "$record" > /dev/null
  else
    cf -X POST "$API" --json "$record" > /dev/null
  fi
  echo "$name -> $IP"
done
