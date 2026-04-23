#!/bin/bash
set -euo pipefail

# Process nginx config templates.
readonly nginx_template_dir="/etc/nginx/templates"
readonly nginx_config_dir="/etc/nginx/conf.d"
for template_file in "${nginx_template_dir}"/*.template; do
  # Skip if no .template files exist.
  [ -e "${template_file}" ] || continue

  template_name=$(basename "${template_file}" .template)
  output_file="${nginx_config_dir}/${template_name}"

  echo "Processing nginx config template: ${template_file} -> ${output_file}"

  # Extract all ${VAR}-style variables from the template so we can
  # tell envsubst which vars to substitute.
  vars=$(grep -oE '\$\{[A-Z_][A-Z_0-9]*\}' "${template_file}" | sort -u | tr '\n' ' ')

  # Substitute env vars in the nginx config template and write the final config.
  envsubst "${vars}" < "${template_file}" > "${output_file}"
done

# Start the backend.
PORT="${BACKEND_PORT}" /serve &

# Start nginx to serve the frontend and proxy queries to the backend.
nginx -g 'daemon off;' &

# Wait for all processes to exit. Starting processes in the background and
# then waiting allows signals to be handled properly.
wait
