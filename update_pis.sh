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

update_device() {
  if [ "$#" -ne 2 ]; then
    echo "expected two arguments"
    return 2
  fi

  readonly device="$1"
  if [ -z "${device}" ]; then
    echo "expected device to be non-empty"
    return 2
  fi

  readonly arch="$2"
  if [ -z "${arch}" ]; then
    echo "expected arch to be non-empty"
    return 2
  fi

  readonly new_binary="${binary_name}_${arch}_${commit}"

  echo -e "\n#################################################"
  echo "# ${device}" "${arch}"
  echo "#################################################"

  ssh "${device}" "sudo systemctl stop ${service_name}.service"

  scp "out/${arch}/${binary_name}" "${device}:~/${new_binary}"
  ssh "${device}" "rm -f ${binary_name} && ln -s ${new_binary} ${binary_name}"

  # Delete the old JWT because it can cause connection problems on restart
  ssh "${device}" 'rm -f ~/.iotcorelogger/iotcorelogger.jwt'

  ssh "${device}" "sudo systemctl start ${service_name}.service"
}

for d in "${!devices[@]}"; do
  arch="${devices[$d]}"

  update_device "${d}" "${arch}" &
done

wait
