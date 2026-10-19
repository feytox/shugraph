#!/bin/sh
# Prepends "[GSD-N] " from the branch name to the commit subject.

msg_file="$1"

branch=$(git symbolic-ref --short HEAD 2>/dev/null) || exit 0
printf '%s' "$branch" | grep -Eq '^GSD-[0-9]+$' || exit 0

subject=$(sed '/^# -* >8 -*$/,$d' "$msg_file" | grep -v '^#' | grep -m 1 '[^[:space:]]')
[ -z "$subject" ] && exit 0
printf '%s' "$subject" | grep -Eq '^(\[GSD-[0-9]+\]|fixup! |squash! |amend! |Merge )' && exit 0

tmp=$(mktemp)
SUBJECT="$subject" PREFIX="[$branch] " awk '
  !done && $0 == ENVIRON["SUBJECT"] { $0 = ENVIRON["PREFIX"] $0; done = 1 }
  { print }
' "$msg_file" > "$tmp" && mv "$tmp" "$msg_file"
