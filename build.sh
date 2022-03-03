#!/usr/bin/env bash

# Allow execution from anywhere to behave as if executed within the directory
# this script resides inside of.
basepath=$(dirname $(realpath $0))
cd $basepath

# Set much more strict behavior for failed commands and unexpanded variables
set -eu

# Allow's the script to refer to itself
me=$(basename $0)

package=bastion.tar.gz

log() {
    echo "[$me] $@"
}

log-warn() {
    echo "[$me] 🟡 $@"
}

log-fatal() {
    echo "[$me] 🔴 $@"
    exit 1
}

# require() accepts a list of space separated commands and checks if they are
# available to the shell. If all commands are available, returns successfully.
# If any of the commands are unavailable, the script returns with a value of 1.
require() {
    missing=""
    for cmd in $@; do
        command -v "$cmd" >/dev/null 2>&1 || { missing="$missing $cmd"; }
    done

    # Even if only one pre-requisite is missing, exit since none are optional
    if [[ "$missing" != "" ]]; then
        log-fatal "please install required commands:$missing"
        exit 1
    fi
}

build() {
    log "🔨 building…"
    go build $basepath
}

clean() {
    log "🧹 cleaning…"
    go clean $basepath
    rm -f $package
}

package() {
    log "📦 packaging…"
    files=("bastion" "docs" "www.example.com")
    if ! tar -cf - "$files" -P | pv -s $(du -sb "$files" | awk '{print $1}') | pigz > "$package"; then
        log-fatal "couldn't package $files"
    fi
}

usage() {
    echo "Usage: $me [all|build|clean|package|help]"
}

main() {
    subcommand=${1-all}
    args=${@:2}

    case $subcommand in
        build)
            build $args
            ;;
        clean)
            clean $args
            ;;
        package)
            package $args
            ;;
        all)
            build
            package
            ;;
        -h|--help|help)
            usage
            ;;
        *)
            log "'$subcommand' is not a valid subcommand."
            usage
            exit 1
            ;;
    esac
}

require \
    go \
    pigz \
    pv \
    realpath \
    tar

main $@
