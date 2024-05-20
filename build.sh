#!/usr/bin/env bash

# Run this script inside of the directory it resides in.
cd "$(dirname "$(realpath "$0")")"
# Set much more strict behavior for failed commands and unexpanded variables.
set -eu
# Allow's the script to refer to itself.
me="$(basename "$0")"

# Text colors
FMT_NONE="$(tput sgr0)"
FMT_RED="$(tput setaf 1)"
FMT_GREEN="$(tput setaf 2)"
FMT_YELLOW="$(tput setaf 3)"

package_files=("bastion" "docs" "LICENSE" "www.example.com")
package="bastion-$(cat VERSION.txt).tar.gz"

function log() {
	echo "[${FMT_GREEN}${me}:info${FMT_NONE}] $*"
}

function log-warn() {
	echo "[${FMT_YELLOW}${me}:warn${FMT_NONE}] $*"
}

function log-fatal() {
	echo "[${FMT_RED}${me}:fatal${FMT_NONE}] $*"
	exit 1
}

# require() accepts a list of space separated commands and checks if they are
# available to be run.
function require() {
	local missing=""
	for cmd in "$@"; do
		command -v "$cmd" >/dev/null 2>&1 || { missing="$missing $cmd"; }
	done

	# Even if only one pre-requisite is missing, exit since none are optional
	if [[ "$missing" != "" ]]; then
		log-fatal "please install required commands:$missing"
	fi
}

function build() {
	log "🔨 building…"
	go build -ldflags "-X github.com/toddgaunt/bastion/internal/errors.ModulePrefix=$(realpath .)/" ./cmd/bastion
}

function clean() {
	log "🧹 cleaning…"
	go clean
	rm -f $package
}

function package() {
	local progress=0
	while getopts ":hp" opt; do
		case ${opt} in
			h)
				echo "$me package [-h|-p]"
				echo "    -h      Display this help message"
				echo "    -p      Show progress bar"
				exit 0
				;;
			p)
				progress=1
				;;
			\?)
				echo "Invalid option: -$OPTARG" 1>&2
				exit 1
				;;
		esac
	done
	shift $((OPTIND -1))

	log "📦 packaging…"
	local size
	size=$(du -s -c -k "${package_files[@]}" | awk '{print $1}' | tail -n 1)

	if [ $progress -eq 1 ]; then
		if ! tar -c -f - "${package_files[@]}" |
			pv -pert -s "${size}k" |
			pigz > "$package"; then
			log-fatal "couldn't package ${package_files[*]}"
		fi
	else
		tar -c -f "$package" "${package_files[@]}"
	fi
}

function usage() {
	echo "Usage: $me [all|clean|package|help]"
}

function main() {
	local subcommand=${1-all}
	local args=${*:2}

	case $subcommand in
		all)
			build "$args"
			;;
		clean)
			clean "$args"
			;;
		package)
			package "$args"
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

main "$@"
