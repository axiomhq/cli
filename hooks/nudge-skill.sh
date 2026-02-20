#!/usr/bin/env bash
# Nudge agents to use Axiom CLI skills instead of running axiom commands
# directly without the skill's guidance.
#
# Uses PostToolUse additionalContext to inject a nudge after the command runs.

input=$(cat)

# Check if the bash command involves axiom query/dataset/stream.
command=$(echo "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)
if [[ -z "$command" ]]; then
  exit 0
fi
if ! echo "$command" | grep -qE 'axiom (query|dataset|stream)'; then
  exit 0
fi

# Only nudge once per session. Extract session_id from hook input, fall back
# to PPID if unavailable (e.g. older Claude Code versions).
session_id=$(echo "$input" | jq -r '.session_id // empty' 2>/dev/null)
marker="${TMPDIR:-/tmp}/.axiom-cli-skill-nudge-${session_id:-$PPID}"
[ -f "$marker" ] && exit 0
touch "$marker"

nudge='<system-reminder>Axiom CLI skills provide guided workflows. Use /axiom-apl for APL query reference, /explore-dataset to discover schema and patterns, /find-traces to analyze OpenTelemetry traces, or /detect-anomalies for statistical anomaly detection.</system-reminder>'

jq -n --arg nudge "$nudge" '{
  hookSpecificOutput: {
    hookEventName: "PostToolUse",
    additionalContext: $nudge
  }
}'
