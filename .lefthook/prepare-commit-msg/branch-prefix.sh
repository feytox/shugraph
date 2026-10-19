#!/bin/sh
# Prepends "[GSD-N] " from the branch name to the commit message.

msg_file="$1"
source="$2"

[ "$source" = "merge" ] && exit 0

branch=$(git symbolic-ref --short HEAD 2>/dev/null) || exit 0
printf '%s' "$branch" | grep -Eq '^GSD-[0-9]+$' || exit 0

head -n 1 "$msg_file" | grep -Eq '^\[GSD-[0-9]+\]' && exit 0

tmp=$(mktemp)
{ printf '[%s] ' "$branch"; cat "$msg_file"; } > "$tmp" && mv "$tmp" "$msg_file"
