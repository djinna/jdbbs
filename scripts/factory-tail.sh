#!/usr/bin/env bash
# factory-tail — watch the factory while someone is working through it.
#
#   scripts/factory-tail.sh            follow live
#   scripts/factory-tail.sh -S -2h     everything since two hours ago, then follow
#
# Shows: client logins, uploads, Inspect runs, transmittal saves, builds,
# pass fulfilment, store checkouts, emails, every WARN/ERROR, and every
# non-GET / failed HTTP request (from srv/reqlog.go). Hides store polling
# noise and successful GETs (those are never logged).
set -euo pipefail
pattern='client login|manuscript uploaded|inspect run|transmittal saved|book conversion|epub generation|build|factory pass|store: (checkout|fulfilled|created)|email sent|WARN|ERROR|INFO http'
journalctl -u prodcal -f -o short-iso --no-hostname "$@" \
  | grep --line-buffered -E "$pattern" \
  | grep --line-buffered -vE 'store: poll' \
  | sed -u -E \
      -e 's/ prodcal\[[0-9]+\]: [0-9\/]+ [0-9:]+ ?//' \
      -e $'s/ ERROR /\e[31m ERROR \e[0m/' \
      -e $'s/ WARN /\e[33m WARN \e[0m/' \
      -e $'s/(who=[^ ]+)/\e[36m\\1\e[0m/' \
      -e $'s/(status=[45][0-9][0-9])/\e[31m\\1\e[0m/'
