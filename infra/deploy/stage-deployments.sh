#!/usr/bin/env bash
set -euo pipefail

PR=${1:?usage: $0 <pr>}

gh api --paginate "repos/{owner}/{repo}/deployments?environment=stage&per_page=100" \
  --jq ".[] | select(.payload.pr == $PR) | .id" |
  while read -r id; do
    state=$(gh api "repos/{owner}/{repo}/deployments/$id/statuses?per_page=1" --jq '.[0].state')
    if [[ $state != inactive ]]; then echo "$id $state"; fi
  done
