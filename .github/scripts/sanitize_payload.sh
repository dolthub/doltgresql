#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

# Inputs via environment variables
RAW_DEP=${RAW_DEP:-}
RAW_SHA=${RAW_SHA:-}
RAW_USER=${RAW_USER:-}
RAW_MAIL=${RAW_MAIL:-}

# --- Validate dependency via allow-list and map to module path + label
case "${RAW_DEP:-}" in
  dolt)
    MODULE='github.com/dolthub/dolt/go'
    LABEL='dolt-bump'
    ;;
  *)
    echo "Unsupported dependency '${RAW_DEP:-}'" >&2
    exit 1
    ;;
esac

# --- Validate head SHA/tag (conservative)
# allow only hex SHAs or safe tag-ish: letters, digits, dot, dash, underscore, plus
if [ -z "${RAW_SHA:-}" ] || ! printf '%s' "$RAW_SHA" | grep -qE '^[A-Za-z0-9._+-]+$'; then
  echo "Invalid head_commit_sha" >&2
  exit 1
fi

# Keep a short 8-char form if it's a hex SHA; otherwise derive short safe token
if printf '%s' "$RAW_SHA" | grep -qiE '^[0-9a-f]{40}$'; then
  SHORT_SHA="${RAW_SHA:0:8}"
else
  SHORT_SHA="$(printf '%s' "$RAW_SHA" | tr -cd 'A-Za-z0-9._+-' | cut -c1-12)"
fi

# --- Determine candidate assignee (map bot -> zachmu), then validate
if [ "${RAW_USER:-}" = "github-actions[bot]" ]; then
  USER_CAND='zachmu'
else
  USER_CAND="${RAW_USER:-}"
fi

if [ -z "${USER_CAND:-}" ] || ! printf '%s' "$USER_CAND" | grep -qE '^[A-Za-z0-9-]{1,39}$'; then
  echo "Invalid assignee username" >&2
  exit 1
fi

# --- Validate email; if invalid, fall back to GitHub noreply
if [ -n "${RAW_MAIL:-}" ] && printf '%s' "$RAW_MAIL" | grep -qE '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'; then
  SAFE_EMAIL="$RAW_MAIL"
else
  SAFE_EMAIL="${USER_CAND}+noreply@users.noreply.github.com"
fi

# --- Build a safe branch name: <assignee>-<short>
BRANCH_NAME="$(printf '%s-%s' "$USER_CAND" "$SHORT_SHA" | tr -cd 'A-Za-z0-9._-')"

# Expose sanitized values as step outputs
{
  echo "label=$LABEL"
  echo "safe_module=$MODULE"
  echo "safe_head=$RAW_SHA"
  echo "safe_assignee=$USER_CAND"
  echo "safe_email=$SAFE_EMAIL"
  echo "safe_branch=$BRANCH_NAME"
  echo "safe_short=$SHORT_SHA"
} >> "${GITHUB_OUTPUT}"
