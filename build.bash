#!/bin/bash
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
readonly script_dir
readonly app=magehelper
readonly cache_volume="go-cache-${app}"
readonly golang=docker.io/library/golang:1.23.1-alpine

readonly cache_path=/go-cache

args=$(getopt --options uc --longoptions update,clear --name "$(basename "$0")" -- "$@")
eval set -- "${args}"

update=false
while :; do
    case "$1" in
        -u | --update)
            readonly update=true
            ;;
        -c | --clear)
            # status 1 means a volume didn't exist, which is fine.
            podman volume rm "${cache_volume}" || test $? = 1
            exit
            ;;
        --)
            shift
            break
            ;;
    esac
    shift
done
readonly update

g() {
    local args
    args=(
        --interactive
        --rm
        --volume "${script_dir}:/src:rw"
        --volume "${cache_volume}:${cache_path}:rw"
        --env GOBIN="${cache_path}/bin"
        --env GOCACHE="${cache_path}/go"
        --env GOMODCACHE="${cache_path}/mod"
        --env CGO_ENABLED=0
        --label app="${app}"
        --label role=builder
        --workdir /src
        "${golang}"
    )
    (set -x; podman run "${args[@]}" "$@")
}

volume_args=(
    --label app="${app}"
    --label role=cache
)

if ! podman volume exists "${cache_volume}"; then
    podman volume create "${volume_args[@]}" "${cache_volume}"
fi

g sh -x <<END
set -euo pipefail
if ${update}; then
    find -name go.mod -exec /bin/sh -c 'cd \$(dirname {}) && go get -u' ';'
fi
find -name go.mod -exec /bin/sh -c 'cd \$(dirname {}) && go mod tidy -go 1.23' ';'

(cd magefiles && go build -o ../bin/mage mage.go)
bin/mage all
END

# vim: et sw=4 ts=4
