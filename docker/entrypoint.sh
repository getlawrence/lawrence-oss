#!/bin/sh
set -eu

# Allow operators to supply a backend URL at runtime without rebuilding the UI bundle.
# When BACKEND_URL (or VITE_BACKEND_URL for backwards compatibility) is set,
# replace the placeholder token inside the generated runtime config file.

CONFIG_PATH="/app/ui/dist/lawrence-config.js"
CONFIG_TEMPLATE_PATH="/app/ui/dist/lawrence-config.js.template"
RUNTIME_URL="${BACKEND_URL:-${VITE_BACKEND_URL:-}}"

if [ -f "$CONFIG_PATH" ] && [ ! -f "$CONFIG_TEMPLATE_PATH" ]; then
  cp "$CONFIG_PATH" "$CONFIG_TEMPLATE_PATH"
fi

if [ -f "$CONFIG_TEMPLATE_PATH" ]; then
  cp "$CONFIG_TEMPLATE_PATH" "$CONFIG_PATH"
fi

if [ -n "$RUNTIME_URL" ] && [ -f "$CONFIG_PATH" ]; then
  # Escape characters that sed treats as special.
  ESCAPED_URL=$(printf '%s\n' "$RUNTIME_URL" | sed 's/[\/&]/\\&/g')
  sed -i "s|__LAWRENCE_BACKEND_URL__|$ESCAPED_URL|g" "$CONFIG_PATH"
fi

exec "$@"

