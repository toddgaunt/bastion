#!/usr/bin/env bash

# Run this script inside of the directory it resides in.
cd $(dirname $(realpath $0))
# Set much more strict behavior for failed commands and unexpanded variables.
set -eu
# Allow's the script to refer to itself.
me=$(basename $0)

# Text colors
FMT_NONE="$(tput sgr0)"
FMT_RED="$(tput setaf 1)"
FMT_GREEN="$(tput setaf 2)"
FMT_YELLOW="$(tput setaf 3)"

package_files=("bastion" "docs" "www.example.com")
package=bastion.tar.gz

log() {
	echo "[$me] $@"
}

log-warn() {
	echo "[${FMT_YELLOW}${me}${FMT_NONE}] $@"
}

log-fatal() {
	echo "[${FMT_RED}${me}${FMT_NONE}] $@"
	exit 1
}

# require() accepts a list of space separated commands and checks if they are
# available to be run. If all commands are available, returns with a value of
# 0. If any of the commands are unavailable, the script returns with a value of
# 1.
require() {
	local missing=""
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
	go build
}

clean() {
	log "🧹 cleaning…"
	go clean
	rm -f $package
}

package() {
	log "📦 packaging…"
	if ! tar -cf - "$package_files" -P |
		pv -s $(du -sb "$package_files" | awk '{print $1}') |
		pigz > "$package"; then
		log-fatal "couldn't package $files"
	fi
}

usage() {
	echo "Usage: $me [all|build|clean|package|help]"
}

main() {
	local subcommand=${1-all}
	local args=${@:2}

	case $subcommand in
		all)
			build $args
			;;
		clean)
			clean $args
			;;
		package)
			package $args
			;;
		-h|--help|help)
			usage
			;;
		*)
			log "$subcommand is not a valid subcommand."
			usage
			exit 1
			;;
	esac
}

require \
	tput \
	go \
	pigz \
	pv \
	realpath \
	tar

main $@
