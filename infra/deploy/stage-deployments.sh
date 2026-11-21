#!/usr/bin/env bash
set -euo pipefail

FILTER=${1:+"select(.payload.pr == $1) |"}

gh api --paginate "repos/{owner}/{repo}/deployments?environment=stage&per_page=100" \
  --jq ".[] | $FILTER \"\(.id) \(.payload.pr)\"" |
  while read -r id pr; do
    state=$(gh api "repos/{owner}/{repo}/deployments/$id/statuses?per_page=1" --jq '.[0].state')
    if [[ $state != inactive ]]; then echo "$id $state $pr"; fi
  done
