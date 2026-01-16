#!/bin/bash

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
  echo "Usage: $0 <semver>"
  exit 1
fi

is_semver() {
  local v="$1"
  # SemVer regex (based on semver.org): MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
  # https://regex101.com/r/vkijKf/1/
  local re='^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*))?(\+([0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*))?$'

  if [[ $v =~ $re ]]; then
    return 0
  else
    return 1
  fi
}

SEMVER="${1}"

if ! is_semver "${SEMVER}"; then
  echo "Error: invalid semver: ${SEMVER}" >&2
  exit 1
fi

git push origin --delete "${SEMVER}"
git tag --delete "${SEMVER}"
git tag "${SEMVER}" --message "Release ${SEMVER}."
git push origin
git push origin "${SEMVER}"
