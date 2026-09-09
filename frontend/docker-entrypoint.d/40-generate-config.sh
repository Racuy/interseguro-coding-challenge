#!/bin/sh
set -eu

: "${API_BASE_URL:=}"

cat > /usr/share/nginx/html/config.js <<EOF
window.APP_CONFIG = { apiBase: "${API_BASE_URL}" };
EOF
