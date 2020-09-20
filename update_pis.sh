#!/bin/bash
set -euo pipefail

# Devices maps Raspberry Pi IP address or name (if set in ssh config) to
# architecture. Binaries of the given architecture will be pushed to that pi.
declare -rA devices=(
  ["rpi"]="armv7"
  ["192.168.1.35"]="armv6"
)

readonly binary_name="iotcorelogger"
readonly service_name="iotcorelogger"

readonly commit="$(git rev-parse --verify --short HEAD)"

echo "Updating to binaries at commit ${commit}"

for d in "${!devices[@]}"; do
  arch="${devices[$d]}"

  new_binary="${binary_name}_${arch}_${commit}"

  echo -e "\n#################################################"
  echo "# ${d}"
  echo "#################################################"

  ssh "${d}" "sudo systemctl stop ${service_name}.service"

  scp "out/${arch}/${binary_name}" "${d}:~/${new_binary}"
  ssh "${d}" "rm -f ${binary_name} && ln -s ${new_binary} ${binary_name}"

  ssh "${d}" "sudo systemctl start ${service_name}.service"
done
