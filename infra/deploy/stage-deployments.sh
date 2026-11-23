#!/usr/bin/env bash
set -euo pipefail

FILTER=${1:+"select(.payload.pr == $1) |"}

gh api graphql --paginate -F owner='{owner}' -F name='{repo}' -f query='
  query($owner: String!, $name: String!, $endCursor: String) {
    repository(owner: $owner, name: $name) {
      deployments(environments: ["stage"], first: 100, after: $endCursor) {
        nodes { databaseId payload latestStatus { state } }
        pageInfo { hasNextPage endCursor }
      }
    }
  }' \
  --jq ".data.repository.deployments.nodes[] | .payload |= (fromjson | fromjson) | $FILTER
    select(.latestStatus.state != \"INACTIVE\") |
    \"\(.databaseId) \(.latestStatus.state // \"null\" | ascii_downcase) \(.payload.pr)\""
