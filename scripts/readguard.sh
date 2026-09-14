#!/bin/bash
# readguard.sh — refuse to dump big files into an agent's context.
#
# Sourced from ~/.profile so it applies to every 'bash --login -c' the Shelley
# bash tool runs (lead and subagents alike). Not a Shelley hook: Shelley has no
# pre-tool-call hook, and this sits on the real execution path anyway.
#
# Policy (docs/CONTEXT-HYGIENE.md): never 'cat' a file over READGUARD_LINES
# lines or READGUARD_BYTES bytes. Use sed -n / rg -n / head / tail, or delegate
# a bulk read to a cheap-model subagent.
#
# Bypass when you really mean it:   'command cat FILE'   or   'READGUARD=off cat FILE'

: "${READGUARD_LINES:=350}"
: "${READGUARD_BYTES:=24000}"

_readguard_check() {
    # $1 = command name, rest = args. Returns 1 (and prints why) if any file
    # argument is too big. Flags and nonexistent paths are ignored.
    local cmd=$1; shift
    [ "${READGUARD:-on}" = off ] && return 0
    local f lines bytes bad=0
    for f in "$@"; do
        case "$f" in -*) continue;; esac
        [ -f "$f" ] || continue
        bytes=$(stat -c %s -- "$f" 2>/dev/null || echo 0)
        lines=$(wc -l < "$f" 2>/dev/null || echo 0)
        if [ "$lines" -gt "$READGUARD_LINES" ] || [ "$bytes" -gt "$READGUARD_BYTES" ]; then
            printf "readguard: refusing '%s %s' (%s lines, %s bytes > %s/%s).\n" \
                "$cmd" "$f" "$lines" "$bytes" "$READGUARD_LINES" "$READGUARD_BYTES" >&2
            bad=1
        fi
    done
    if [ $bad = 1 ]; then
        printf "readguard: use 'sed -n A,Bp FILE', 'rg -n PATTERN FILE', head/tail, or delegate the\n" >&2
        printf "readguard: bulk read to a cheap subagent. Bypass: 'command %s FILE' or READGUARD=off.\n" "$cmd" >&2
        return 1
    fi
    return 0
}

cat()  { _readguard_check cat  "$@" && command cat  "$@"; }
less() { _readguard_check less "$@" && command less "$@"; }
more() { _readguard_check more "$@" && command more "$@"; }
